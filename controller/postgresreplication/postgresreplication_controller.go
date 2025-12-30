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

package postgresreplication

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	composev1alpha1 "github.com/upmio/compose-operator/api/v1alpha1"
	"github.com/upmio/compose-operator/pkg/k8sutil"
	"github.com/upmio/compose-operator/pkg/postgresutil"
)

const (
	ErrSynced    = "ErrSynced"
	Synced       = "Synced"
	requeueAfter = 10 * time.Second

	appName = "postgres-replication"

	SyncTopologyFailed  = "SyncTopologyFailed"
	SyncTopologySucceed = "SyncTopologySucceed"

	SyncResourceFailed  = "SyncResourceFailed"
	SyncResourceSucceed = "SyncResourceSucceed"

	readOnlyKey = "compose-operator/postgres-replication.readonly"
	defaultKey  = "compose-operator/postgres-replication.name"

	portName      = "postgresql"
	containerName = "postgresql"

	desiredUnitStart = "start"
	desiredUnitStop  = "stop"
	signalFileName   = "standby.signal"
	cmdPrefix        = "/bin/sh"
)

// ReconcilePostgresReplication reconciles a PostgresReplication object
type ReconcilePostgresReplication struct {
	client         client.Client
	scheme         *runtime.Scheme
	recorder       record.EventRecorder
	logger         logr.Logger
	execer         k8sutil.IExec
	unitController k8sutil.IUnitControl
}

// blank assignment to verify that ReconcilePostgresReplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePostgresReplication{}

// syncCtx
type syncContext struct {
	instance  *composev1alpha1.PostgresReplication
	admin     postgresutil.IAdmin
	ctx       context.Context
	reqLogger logr.Logger
}

// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=postgresreplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=postgresreplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=postgresreplications/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;update;patch;list;watch
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=units,verbs=get;list;update;patch
// +kubebuilder:rbac:groups=upm.syntropycloud.io,resources=units/status,verbs=get

// Reconcile reconcile postgres replication
func (r *ReconcilePostgresReplication) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.logger.WithValues("namespace", req.Namespace, "name", req.Name)
	reqLogger.Info("starting reconciliation for postgres replication")
	startTime := time.Now()
	defer func() {
		reqLogger.Info("finished reconciliation", "duration", time.Since(startTime))
	}()

	// Fetch the MysqlReplication instance
	instance := &composev1alpha1.PostgresReplication{}
	if err := r.client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("postgres replication resource not found, probably deleted.")
			return reconcile.Result{}, nil
		}

		reqLogger.Error(err, "failed to fetch postgres replication resource")
		return reconcile.Result{}, err
	}

	//check postgres replication instance is valid
	if len(instance.Spec.Standby) == 0 && instance.Spec.Mode == composev1alpha1.PostgresRplSync {
		r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, "sync mode requires at least one standby")

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

	// Get the password of PostgreSQL user for connection and replication.
	password, replicationPassword, err := decryptSecret(r.client, reqLogger, instance)
	if err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncTopologyCondition(err))
		r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
	} else {
		// Initialize the PostgreSQL Admin
		admin := newPostgresAdmin(instance, password, reqLogger)
		defer admin.Close()

		if value, found := instance.GetAnnotations()[composev1alpha1.SkipReconcileKey]; !found || value == "false" {
			syncCtx.admin = admin
			if err := r.handlePostgresReplicationInstance(syncCtx, replicationPassword, password); err != nil {
				meta.SetStatusCondition(&instance.Status.Conditions, newFailedSyncTopologyCondition(err))
				r.recorder.Event(instance, corev1.EventTypeWarning, ErrSynced, err.Error())
			} else {
				meta.SetStatusCondition(&instance.Status.Conditions, newSucceedSyncTopologyCondition())
			}
		}

		generateTopologyStatusByReplicationInfo(admin.GetReplicationStatus(ctx), instance)
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
	r := &ReconcilePostgresReplication{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(appName),
		logger:   ctrl.Log.WithName(appName),
	}

	gvk := runtimeschema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}
	c, err := rest.HTTPClientFor(mgr.GetConfig())
	if err != nil {
		return err
	}
	restClient, err := apiutil.RESTClientForGVK(gvk, false, mgr.GetConfig(), serializer.NewCodecFactory(scheme.Scheme), c)
	if err != nil {
		return err
	}
	r.execer = k8sutil.NewRemoteExec(restClient, mgr.GetConfig(), r.logger)

	if r.unitController, err = k8sutil.NewUnitController(mgr.GetConfig()); err != nil {
		return err
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

		// Allow create events
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},

		// Allow delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},

		// Allow generic events (e.g., external triggers)
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&composev1alpha1.PostgresReplication{}).
		Owns(&corev1.Service{}).
		Watches(&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(podMapFunc),
			builder.WithPredicates(updatePred), // Apply the predicate
		).
		WithOptions(
			controller.Options{MaxConcurrentReconciles: 10},
		).
		Complete(r)
}
