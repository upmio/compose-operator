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

package rediscluster

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	"github.com/upmio/compose-operator/pkg/redisutil"
)

const (
	ErrSynced    = "ErrSynced"
	Synced       = "Synced"
	requeueAfter = 10 * time.Second

	SyncTopologyFailed  = "SyncTopologyFailed"
	SyncTopologySucceed = "SyncTopologySucceed"

	appName = "redis-cluster"
)

// ReconcileRedisCluster reconciles a RedisCluster object
type ReconcileRedisCluster struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	logger   logr.Logger
}

// blank assignment to verify that ReconcileRedisCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRedisCluster{}

type syncContext struct {
	instance  *composev1alpha1.RedisCluster
	admin     redisutil.IClusterAdmin
	reqLogger logr.Logger
	ctx       context.Context
}

// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=redisclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=redisclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=redisclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile reconcile redis cluster
func (r *ReconcileRedisCluster) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.logger.WithValues("namespace", req.Namespace, "name", req.Name)
	reqLogger.Info("starting reconciliation for redis cluster")
	startTime := time.Now()
	defer func() {
		reqLogger.Info("finished reconciliation", "duration", time.Since(startTime))
	}()

	// Fetch the RedisCluster instance
	instance := &composev1alpha1.RedisCluster{}
	if err := r.client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("redis cluster resource not found, probably deleted.")
			return reconcile.Result{}, nil
		}

		reqLogger.Error(err, "failed to fetch redis cluster resource")
		return reconcile.Result{}, err
	}

	syncCtx := &syncContext{
		instance:  instance,
		reqLogger: reqLogger,
		ctx:       ctx,
	}

	oldStatus := instance.Status.DeepCopy()
	instance.Status = buildDefaultTopologyStatus(instance)

	// Get the password of Redis for connection.
	password, err := decryptSecret(r.client, instance, reqLogger)
	if err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncTopologyCondition(err))
		r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
	} else {
		// Initialize the Redis Admin
		admin := newRedisAdmin(instance, password, reqLogger)
		defer admin.Close()

		if value, found := instance.ObjectMeta.GetAnnotations()[composev1alpha1.SkipReconcileKey]; !found || value == "false" {
			syncCtx.admin = admin
			if err := r.handleRedisClusterInstance(syncCtx); err != nil {
				meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncTopologyCondition(err))
				r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
			} else {
				meta.SetStatusCondition(&instance.Status.Conditions, newSucceedSyncTopologyCondition())
			}
		}

		info, _ := admin.GetClusterInfos()

		generateTopologyStatusByReplicationInfo(info, instance)
	}

	r.updateInstanceIfNeed(instance, oldStatus, reqLogger)

	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: requeueAfter,
	}, nil
}

func Setup(mgr ctrl.Manager) error {
	r := &ReconcileRedisCluster{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(appName),
		logger:   ctrl.Log.WithName(appName),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&composev1alpha1.RedisCluster{}).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 10},
		).
		Complete(r)
}
