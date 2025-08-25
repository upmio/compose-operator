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

package mysqlgroupreplication

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/mysqlutil"
)

const (
	appName      = "mysql-group-replication"
	ErrSynced    = "ErrSynced"
	Synced       = "Synced"
	requeueAfter = 10 * time.Second

	SyncTopologyFailed  = "SyncTopologyFailed"
	SyncTopologySucceed = "SyncTopologySucceed"

	SyncResourceFailed  = "SyncResourceFailed"
	SyncResourceSucceed = "SyncResourceSucceed"

	defaultKey = "compose-operator.mysqlgroupreplication.name"
)

// ReconcileMysqlGroupReplication reconciles a MysqlGroupReplication object
type ReconcileMysqlGroupReplication struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	logger   logr.Logger
}

// blank assignment to verify that ReconcileMysqlGroupReplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMysqlGroupReplication{}

// syncCtx
type syncContext struct {
	instance  *composev1alpha1.MysqlGroupReplication
	admin     mysqlutil.IGroupAdmin
	ctx       context.Context
	reqLogger logr.Logger
}

// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=mysqlgroupreplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=mysqlgroupreplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=mysqlgroupreplications/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile reconcile mysql group replication
func (r *ReconcileMysqlGroupReplication) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.logger.WithValues("namespace", req.Namespace, "name", req.Name)
	reqLogger.Info("starting reconciliation for mysql group replication")
	startTime := time.Now()
	defer func() {
		reqLogger.Info("finished reconciliation", "duration", time.Since(startTime))
	}()

	// Fetch the MysqlGroupReplication instance
	instance := &composev1alpha1.MysqlGroupReplication{}
	if err := r.client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("mysql group replication resource not found, probably deleted.")
			return reconcile.Result{}, nil
		}

		reqLogger.Error(err, "failed to fetch mysql group replication resource")
		return reconcile.Result{}, err
	}

	//check mysql group replication instance is valid
	groupMemberCount := len(instance.Spec.Member)
	if groupMemberCount == 1 || groupMemberCount%2 == 0 {
		r.recorder.Eventf(instance, corev1.EventTypeWarning, ErrSynced, "invalid mysql group replication member count '%d'", groupMemberCount)

		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: requeueAfter,
		}, nil
	}

	syncCtx := &syncContext{
		instance:  instance,
		reqLogger: reqLogger,
		ctx:       ctx,
	}

	oldStatus := instance.Status.DeepCopy()
	instance.Status = buildDefaultTopologyStatus(instance)

	//Get the password of MySQL user for connection and replication.
	password, replicationPassword, err := decryptSecret(r.client, instance)
	if err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncTopologyCondition(err))
		r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
	} else {
		// Initialize the MySQL GroupAdmin
		admin := newMysqlAdmin(instance, password, reqLogger)
		defer admin.Close()

		if value, found := instance.GetAnnotations()[composev1alpha1.SkipReconcileKey]; !found || value == "false" {
			syncCtx.admin = admin
			if err := r.handleMysqlGroupReplicationInstance(syncCtx, replicationPassword); err != nil {
				meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncTopologyCondition(err))
				r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
			} else {
				meta.SetStatusCondition(&instance.Status.Conditions, newSucceedSyncTopologyCondition())
			}
		}
		generateTopologyStatusByReplicationInfo(admin.GetGroupInfos(ctx, groupMemberCount), instance)
	}

	if err := r.handleResources(syncCtx); err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncResourceCondition(err))
		instance.Status.Ready = false
		r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
	} else {
		meta.SetStatusCondition(&instance.Status.Conditions, newSucceedSyncResourceCondition())
	}

	r.updateInstanceIfNeed(ctx, instance, oldStatus, reqLogger)

	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: requeueAfter,
	}, nil
}

func Setup(mgr ctrl.Manager) error {
	r := &ReconcileMysqlGroupReplication{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(appName),
		logger:   ctrl.Log.WithName(appName),
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, fmt.Sprintf("%s.%s", "metadata.labels.", defaultKey), func(o client.Object) []string {
		pod := o.(*corev1.Pod)
		if val, ok := pod.Labels[defaultKey]; ok {
			return []string{val}
		}
		return nil
	}); err != nil {
		return err
	}

	// Predicate to trigger reconciliation only on size changes in the Busybox spec
	updatePred := predicate.Funcs{
		// Only allow updates when the spec.size of the Busybox resource changes
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObj := e.ObjectOld.(*corev1.Pod)
			newObj := e.ObjectNew.(*corev1.Pod)

			// Trigger reconciliation only if the spec.size field has changed
			return oldObj.Status.Phase != newObj.Status.Phase
		},

		// Allow to create events
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},

		// Allow to delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},

		// Allow to generic events (e.g., external triggers)
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&composev1alpha1.MysqlGroupReplication{}).
		Watches(&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(podMapFunc),
			builder.WithPredicates(updatePred), // Apply the predicate
		).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 10},
		).
		Complete(r)
}
