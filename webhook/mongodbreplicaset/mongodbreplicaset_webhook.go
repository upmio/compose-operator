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

package mongodbreplicaset

import (
	"context"
	"github.com/upmio/compose-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mongodbreplicasetlog = ctrl.Log.WithName("mongodb-replicaset").WithValues("version", "v1alpha1")

type mongodbReplicaSetAdmission struct {
}

// Setup will setup the manager to manage the webhooks
func Setup(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.MongoDBReplicaSet{}).
		WithValidator(&mongodbReplicaSetAdmission{}).
		WithDefaulter(&mongodbReplicaSetAdmission{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-upm-syntropycloud-io-v1alpha1-mongodbreplicaset,mutating=true,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=mongodbreplicasets,verbs=create;update,versions=v1alpha1,name=mmongodbreplicaset.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &mongodbReplicaSetAdmission{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *mongodbReplicaSetAdmission) Default(ctx context.Context, obj runtime.Object) error {
	instance := obj.(*v1alpha1.MongoDBReplicaSet)
	mongodbreplicasetlog.Info("default", "name", instance.Name)

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-upm-syntropycloud-io-v1alpha1-mongodbreplicaset,mutating=false,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=mongodbreplicasets,verbs=create;update,versions=v1alpha1,name=vmongodbreplicaset.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &mongodbReplicaSetAdmission{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *mongodbReplicaSetAdmission) ValidateCreate(ctx context.Context, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.MongoDBReplicaSet)

	mongodbreplicasetlog.Info("validate delete", "name", instance.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *mongodbReplicaSetAdmission) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.MongoDBReplicaSet)

	mongodbreplicasetlog.Info("validate delete", "name", instance.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *mongodbReplicaSetAdmission) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	instance := obj.(*v1alpha1.MongoDBReplicaSet)

	mongodbreplicasetlog.Info("validate delete", "name", instance.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
