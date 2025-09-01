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
	"github.com/upmio/compose-operator/api/v1alpha1"
	"net"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var postgresreplicationlog = ctrl.Log.WithName("postgres-replication").WithValues("version", "v1alpha1")

type postgresReplicationAdmission struct {
}

// Setup will setup the manager to manage the webhooks
func Setup(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.PostgresReplication{}).
		WithValidator(&postgresReplicationAdmission{}).
		WithDefaulter(&postgresReplicationAdmission{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-upm-syntropycloud-io-v1alpha1-postgresreplication,mutating=true,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=postgresreplications,verbs=create;update,versions=v1alpha1,name=mpostgresreplication.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &postgresReplicationAdmission{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *postgresReplicationAdmission) Default(ctx context.Context, obj runtime.Object) error {
	instance := obj.(*v1alpha1.PostgresReplication)
	postgresreplicationlog.Info("default", "name", instance.Name)

	// Set default values for PostgresReplication
	if instance.Spec.Mode == "" {
		instance.Spec.Mode = v1alpha1.PostgresRplAsync
		postgresreplicationlog.Info("setting default mode", "mode", v1alpha1.PostgresRplAsync)
	}

	// Set default secret key names
	if instance.Spec.Secret.Postgresql == "" {
		instance.Spec.Secret.Postgresql = "postgres"
	}
	if instance.Spec.Secret.Replication == "" {
		instance.Spec.Secret.Replication = "replication"
	}

	// Set default service type
	if instance.Spec.Service == nil {
		instance.Spec.Service = &v1alpha1.Service{Type: v1alpha1.ServiceTypeClusterIP}
	} else if instance.Spec.Service.Type == "" {
		instance.Spec.Service.Type = v1alpha1.ServiceTypeClusterIP
	}

	if instance.Spec.AESSecret == nil {
		instance.Spec.AESSecret = &v1alpha1.AESSecret{
			Name: "aes-secret-key",
			Key:  "AES_SECRET_KEY",
		}
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-upm-syntropycloud-io-v1alpha1-postgresreplication,mutating=false,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=postgresreplications,verbs=create;update,versions=v1alpha1,name=vpostgresreplication.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &postgresReplicationAdmission{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *postgresReplicationAdmission) ValidateCreate(ctx context.Context, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.PostgresReplication)
	postgresreplicationlog.Info("validate create", "name", instance.Name)

	return r.validatePostgresReplication(instance)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *postgresReplicationAdmission) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.PostgresReplication)

	postgresreplicationlog.Info("validate update", "name", instance.Name)

	oldPR := oldObj.(*v1alpha1.PostgresReplication)

	// Validate the new object
	warnings, err = r.validatePostgresReplication(instance)
	if err != nil {
		return warnings, err
	}

	// Additional update-specific validations
	var allErrs field.ErrorList

	// Primary node should not be changed without careful consideration
	if oldPR.Spec.Primary != nil && instance.Spec.Primary != nil {
		if oldPR.Spec.Primary.Name != instance.Spec.Primary.Name ||
			oldPR.Spec.Primary.Host != instance.Spec.Primary.Host ||
			oldPR.Spec.Primary.Port != instance.Spec.Primary.Port {
			warnings = append(warnings, "Changing primary node may cause data inconsistency. Ensure proper backup and synchronization.")
		}
	}

	// Secret reference should not change
	if oldPR.Spec.Secret.Name != instance.Spec.Secret.Name {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec").Child("secret").Child("name"),
			instance.Spec.Secret.Name,
			"secret reference cannot be changed after creation",
		))
	}

	if len(allErrs) > 0 {
		return warnings, allErrs.ToAggregate()
	}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *postgresReplicationAdmission) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	instance := obj.(*v1alpha1.PostgresReplication)

	postgresreplicationlog.Info("validate delete", "name", instance.Name)

	// Add warning about potential data loss
	warnings = append(warnings, "Deleting PostgresReplication will stop managing replication but won't affect existing PostgreSQL instances. Ensure proper cleanup if needed.")

	return warnings, nil
}

// validatePostgresReplication performs comprehensive validation of PostgresReplication spec
func (r *postgresReplicationAdmission) validatePostgresReplication(instance *v1alpha1.PostgresReplication) (admission.Warnings, error) {
	var allErrs field.ErrorList
	var warnings admission.Warnings

	// Validate primary node
	if instance.Spec.Primary == nil {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec").Child("primary"),
			"primary node is required",
		))
	} else {
		if err := r.validateCommonNode("primary", instance.Spec.Primary, field.NewPath("spec").Child("primary")); err != nil {
			allErrs = append(allErrs, err...)
		}
	}

	// Validate standby nodes
	if len(instance.Spec.Standby) == 0 {
		warnings = append(warnings, "No standby nodes specified. This will create a single-node setup without replication.")
	} else {
		for i, standby := range instance.Spec.Standby {
			fieldPath := field.NewPath("spec").Child("standby").Index(i)
			if err := r.validateCommonNode("standby", standby, fieldPath); err != nil {
				allErrs = append(allErrs, err...)
			}
		}

		// Check for duplicate node names
		if err := r.validateUniqueNodeNames(instance); err != nil {
			allErrs = append(allErrs, err...)
		}
	}

	// Validate secret configuration
	if err := r.validateSecret(instance); err != nil {
		allErrs = append(allErrs, err...)
	}

	// Validate replication mode
	if instance.Spec.Mode != v1alpha1.PostgresRplAsync && instance.Spec.Mode != v1alpha1.PostgresRplSync {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec").Child("mode"),
			instance.Spec.Mode,
			"mode must be either 'rpl_async' or 'rpl_sync'",
		))
	}

	// Validate service configuration
	if instance.Spec.Service != nil {
		if err := r.validateService(instance); err != nil {
			allErrs = append(allErrs, err...)
		}
	}

	if len(allErrs) > 0 {
		return warnings, allErrs.ToAggregate()
	}

	return warnings, nil
}

// validateCommonNode validates common node fields
func (r *postgresReplicationAdmission) validateCommonNode(nodeType string, node *v1alpha1.CommonNode, fieldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Validate name
	if node.Name == "" {
		allErrs = append(allErrs, field.Required(
			fieldPath.Child("name"),
			fmt.Sprintf("%s node name is required", nodeType),
		))
	}

	// Validate host
	if node.Host == "" {
		allErrs = append(allErrs, field.Required(
			fieldPath.Child("host"),
			fmt.Sprintf("%s node host is required", nodeType),
		))
	} else {
		// Basic host validation (IP address or hostname)
		if net.ParseIP(node.Host) == nil && !isValidHostname(node.Host) {
			allErrs = append(allErrs, field.Invalid(
				fieldPath.Child("host"),
				node.Host,
				"host must be a valid IP address or hostname",
			))
		}
	}

	// Validate port
	if node.Port <= 0 || node.Port > 65535 {
		allErrs = append(allErrs, field.Invalid(
			fieldPath.Child("port"),
			node.Port,
			"port must be between 1 and 65535",
		))
	}

	return allErrs
}

// validateUniqueNodeNames ensures all node names are unique
func (r *postgresReplicationAdmission) validateUniqueNodeNames(instance *v1alpha1.PostgresReplication) field.ErrorList {
	var allErrs field.ErrorList
	nodeNames := make(map[string]bool)

	// Check primary name
	if instance.Spec.Primary != nil && instance.Spec.Primary.Name != "" {
		nodeNames[instance.Spec.Primary.Name] = true
	}

	// Check standby names
	for i, standby := range instance.Spec.Standby {
		if standby.Name != "" {
			if nodeNames[standby.Name] {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath("spec").Child("standby").Index(i).Child("name"),
					standby.Name,
					"node names must be unique across primary and standby nodes",
				))
			}
			nodeNames[standby.Name] = true
		}
	}

	return allErrs
}

// validateSecret validates secret configuration
func (r *postgresReplicationAdmission) validateSecret(instance *v1alpha1.PostgresReplication) field.ErrorList {
	var allErrs field.ErrorList
	secretPath := field.NewPath("spec").Child("secret")

	if instance.Spec.Secret.Name == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("name"),
			"secret name is required",
		))
	}

	if instance.Spec.Secret.Postgresql == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("postgres"),
			"postgres secret key is required",
		))
	}

	if instance.Spec.Secret.Replication == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("replication"),
			"replication secret key is required",
		))
	}

	return allErrs
}

// validateService validates service configuration
func (r *postgresReplicationAdmission) validateService(instance *v1alpha1.PostgresReplication) field.ErrorList {
	var allErrs field.ErrorList
	servicePath := field.NewPath("spec").Child("service")

	validServiceTypes := map[v1alpha1.ServiceType]bool{
		v1alpha1.ServiceTypeClusterIP:    true,
		v1alpha1.ServiceTypeNodePort:     true,
		v1alpha1.ServiceTypeLoadBalancer: true,
		v1alpha1.ServiceTypeExternalName: true,
	}

	if !validServiceTypes[instance.Spec.Service.Type] {
		allErrs = append(allErrs, field.Invalid(
			servicePath.Child("type"),
			instance.Spec.Service.Type,
			"service type must be ClusterIP, NodePort, LoadBalancer, or ExternalName",
		))
	}

	return allErrs
}

// isValidHostname performs basic hostname validation according to DNS specifications.
// This function validates hostnames used for database connections to ensure compatibility
// with DNS resolution and network protocols.
func isValidHostname(hostname string) bool {
	// RFC 1035: DNS names have a maximum length of 255 octets in wire format.
	// Since each label requires a length prefix byte and the name ends with a zero byte,
	// the maximum displayable length is 253 characters (255 - 1 length byte - 1 zero byte).
	// Empty hostnames are also invalid.
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Split hostname into labels (parts separated by dots) and validate each one.
	// DNS labels are the individual components of a domain name (e.g., "postgres", "example", "com").
	for _, part := range strings.Split(hostname, ".") {
		// RFC 1035: Each label must be between 1-63 characters.
		// Empty labels (e.g., "example..com") and labels longer than 63 chars are invalid.
		// The 63-character limit comes from the 6-bit length field in DNS label encoding.
		if len(part) == 0 || len(part) > 63 {
			return false
		}

		// RFC 952 & RFC 1123: Valid hostname characters are letters, digits, and hyphens.
		// This character set ensures compatibility across all DNS implementations and
		// network protocols that may need to resolve this hostname.
		for _, r := range part {
			if (r < 'a' || r > 'z') &&
				(r < 'A' || r > 'Z') &&
				(r < '0' || r > '9') &&
				r != '-' {
				return false
			}
		}

		// RFC 952: Labels cannot start or end with a hyphen.
		// This prevents ambiguity in parsing and ensures compatibility with older systems
		// that might interpret leading/trailing hyphens as command-line options.
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
	}

	return true
}
