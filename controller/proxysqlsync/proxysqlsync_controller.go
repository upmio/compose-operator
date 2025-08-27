/*
Copyright 2025 The Compose Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package proxysqlsync

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
	"github.com/upmio/compose-operator/pkg/proxysqlutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ErrSynced    = "ErrSynced"
	Synced       = "Synced"
	requeueAfter = 10 * time.Second

	SyncServerFailed  = "SyncServerFailed"
	SyncServerSucceed = "SyncServerSucceed"
	SyncUserFailed    = "SyncUserFailed"
	SyncUserSucceed   = "SyncUserSucceed"

	appName = "proxysql-sync"

	ReaderHostGroupId = 1
	WriterHostGroupId = 0
	MaxConnections    = 2000
)

// ReconcileProxysqlSync reconciles a ProxysqlSync object
type ReconcileProxysqlSync struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	logger   logr.Logger
}

// blank assignment to verify that ReconcileProxysqlSync implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileProxysqlSync{}

type syncContext struct {
	instance      *composev1alpha1.ProxysqlSync
	mrInstance    *composev1alpha1.MysqlReplication
	mysqlAdmin    mysqlutil.IReplicationAdmin
	proxysqlAdmin proxysqlutil.IAdmin
	reqLogger     logr.Logger
	ctx           context.Context
}

// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=proxysqlsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=proxysqlsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=proxysqlsyncs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile reconcile proxysql sync
func (r *ReconcileProxysqlSync) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.logger.WithValues("namespace", req.Namespace, "name", req.Name)
	reqLogger.Info("starting reconciliation for proxysql sync")
	startTime := time.Now()
	defer func() {
		reqLogger.Info("finished reconciliation", "duration", time.Since(startTime))
	}()

	// Fetch the ProxysqlSync instance
	instance := &composev1alpha1.ProxysqlSync{}
	if err := r.client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("proxysql sync resource not found, probably deleted.")
			return reconcile.Result{}, nil
		}

		reqLogger.Error(err, "failed to fetch proxysql sync resource")
		return reconcile.Result{}, err
	}

	oldStatus := instance.Status.DeepCopy()
	instance.Status = buildDefaultTopologyStatus(instance)

	// Fetch the MysqlReplication instance
	mrInstance := &composev1alpha1.MysqlReplication{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.MysqlReplication,
		Namespace: instance.Namespace,
	}, mrInstance); err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncUserCondition(err))
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncServerCondition(err))

		r.recorder.Eventf(instance, corev1.EventTypeWarning, ErrSynced, "failed to get mysql replication instance [%s]: %s", instance.Spec.MysqlReplication, err.Error())

		r.updateInstanceIfNeed(ctx, instance, oldStatus, reqLogger)

		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: requeueAfter,
		}, nil
	}

	syncCtx := &syncContext{
		instance:   instance,
		reqLogger:  reqLogger,
		ctx:        ctx,
		mrInstance: mrInstance,
	}

	mysqlPassword, proxysqlPassword, err := decryptSecret(r.client, reqLogger, instance)

	if err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncServerCondition(err))
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncUserCondition(err))
	} else {
		// Initialize the MySQL ReplicationAdmin
		mysqlAdmin := newMysqlAdmin(mrInstance, instance, mysqlPassword, reqLogger)

		defer mysqlAdmin.Close()

		syncCtx.mysqlAdmin = mysqlAdmin

		// Initialize the ProxySQL ReplicationAdmin
		proxysqlAdmin := newProxysqlAdmin(instance, proxysqlPassword, reqLogger)

		defer proxysqlAdmin.Close()

		syncCtx.proxysqlAdmin = proxysqlAdmin

		if value, found := instance.GetAnnotations()[composev1alpha1.SkipReconcileKey]; !found || value == "false" {
			instance.Status.Ready = true
			if err := r.syncMysqlReplicationServer(syncCtx); err != nil {
				instance.Status.Ready = false
				meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncServerCondition(err))
				r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
			} else {
				meta.SetStatusCondition(&instance.Status.Conditions, newSucceedSyncServerCondition())
			}

			if err := r.syncMysqlUser(syncCtx); err != nil {
				instance.Status.Ready = false
				meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncUserCondition(err))
				r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
			} else {
				meta.SetStatusCondition(&instance.Status.Conditions, newSucceedSyncUserCondition())
			}
		}

		generateTopologyStatusByReplicationInfo(proxysqlAdmin, instance)
	}

	r.updateInstanceIfNeed(ctx, instance, oldStatus, reqLogger)

	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: requeueAfter,
	}, nil
}

func Setup(mgr ctrl.Manager) error {
	r := &ReconcileProxysqlSync{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(appName),
		logger:   ctrl.Log.WithName(appName),
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &composev1alpha1.ProxysqlSync{}, ".spec.mysqlReplication", func(rawObj client.Object) []string {
		proxysqlSync := rawObj.(*composev1alpha1.ProxysqlSync)
		if proxysqlSync.Spec.MysqlReplication == "" {
			return nil
		}
		return []string{proxysqlSync.Spec.MysqlReplication}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&composev1alpha1.ProxysqlSync{}).
		Watches(&composev1alpha1.MysqlReplication{},
			handler.EnqueueRequestsFromMapFunc(r.triggerReconcileBecauseMysqlReplicationHasChanged),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 10},
		).
		Complete(r)
}
