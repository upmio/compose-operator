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
var mysqlgroupreplicationlog = ctrl.Log.WithName("mysql-group-replication").WithValues("version", "v1alpha1")

type mysqlGroupReplicationAdmission struct {
}

// Setup will setup the manager to manage the webhooks
func Setup(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.MysqlGroupReplication{}).
		WithValidator(&mysqlGroupReplicationAdmission{}).
		WithDefaulter(&mysqlGroupReplicationAdmission{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-upm-syntropycloud-io-v1alpha1-mysqlgroupreplication,mutating=true,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=mysqlgroupreplications,verbs=create;update,versions=v1alpha1,name=mmysqlgroupreplication.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &mysqlGroupReplicationAdmission{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *mysqlGroupReplicationAdmission) Default(ctx context.Context, obj runtime.Object) error {
	instance := obj.(*v1alpha1.MysqlGroupReplication)
	mysqlgroupreplicationlog.Info("default", "name", instance.Name)

	// Set default secret key names
	if instance.Spec.Secret.Mysql == "" {
		instance.Spec.Secret.Mysql = "mysql"
	}
	if instance.Spec.Secret.Replication == "" {
		instance.Spec.Secret.Replication = "replication"
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-upm-syntropycloud-io-v1alpha1-mysqlgroupreplication,mutating=false,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=mysqlgroupreplications,verbs=create;update,versions=v1alpha1,name=vmysqlgroupreplication.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &mysqlGroupReplicationAdmission{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *mysqlGroupReplicationAdmission) ValidateCreate(ctx context.Context, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.MysqlGroupReplication)
	mysqlgroupreplicationlog.Info("validate create", "name", instance.Name)

	return r.validateMysqlGroupReplication(instance)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *mysqlGroupReplicationAdmission) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.MysqlGroupReplication)

	mysqlgroupreplicationlog.Info("validate update", "name", instance.Name)

	oldMGR := oldObj.(*v1alpha1.MysqlGroupReplication)

	// Validate the new object
	warnings, err = r.validateMysqlGroupReplication(instance)
	if err != nil {
		return warnings, err
	}

	// Additional update-specific validations
	var allErrs field.ErrorList

	// Secret reference should not change
	if oldMGR.Spec.Secret.Name != instance.Spec.Secret.Name {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec").Child("secret").Child("name"),
			instance.Spec.Secret.Name,
			"secret reference cannot be changed after creation",
		))
	}

	// Warn about changing group members
	if len(oldMGR.Spec.Member) != len(instance.Spec.Member) {
		warnings = append(warnings, "Changing group members may cause group reconfiguration and temporary unavailability.")
	}

	if len(allErrs) > 0 {
		return warnings, allErrs.ToAggregate()
	}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *mysqlGroupReplicationAdmission) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	instance := obj.(*v1alpha1.MysqlGroupReplication)

	mysqlgroupreplicationlog.Info("validate delete", "name", instance.Name)

	// Add warning about potential data loss
	warnings = append(warnings, "Deleting MysqlGroupReplication will stop managing group replication but won't affect existing MySQL instances. Ensure proper cleanup if needed.")

	return warnings, nil
}

// validateMysqlGroupReplication performs comprehensive validation of MysqlGroupReplication spec
func (r *mysqlGroupReplicationAdmission) validateMysqlGroupReplication(instance *v1alpha1.MysqlGroupReplication) (admission.Warnings, error) {
	var allErrs field.ErrorList
	var warnings admission.Warnings

	// Validate group members
	if len(instance.Spec.Member) == 0 {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec").Child("member"),
			"at least one group member is required",
		))
	} else {
		// MySQL Group Replication requires at least 3 members for proper fault tolerance
		if len(instance.Spec.Member) < 3 {
			warnings = append(warnings, "MySQL Group Replication requires at least 3 members for proper fault tolerance and automatic failover.")
		}

		// MySQL Group Replication has a maximum of 9 members
		if len(instance.Spec.Member) > 9 {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec").Child("member"),
				len(instance.Spec.Member),
				"MySQL Group Replication supports a maximum of 9 members",
			))
		}

		for i, member := range instance.Spec.Member {
			fieldPath := field.NewPath("spec").Child("member").Index(i)
			if err := r.validateCommonNode("member", member, fieldPath); err != nil {
				allErrs = append(allErrs, err...)
			}
		}

		// Check for duplicate member names
		if err := r.validateUniqueNodeNames(instance); err != nil {
			allErrs = append(allErrs, err...)
		}
	}

	// Validate secret configuration
	if err := r.validateSecret(instance); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) > 0 {
		return warnings, allErrs.ToAggregate()
	}

	return warnings, nil
}

// validateCommonNode validates common node fields
func (r *mysqlGroupReplicationAdmission) validateCommonNode(nodeType string, node *v1alpha1.CommonNode, fieldPath *field.Path) field.ErrorList {
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

	// MySQL default port warning
	if node.Port != 3306 {
		// This would add a warning, but we return field.ErrorList here
		// The warning will be handled at a higher level
	}

	return allErrs
}

// validateUniqueNodeNames ensures all group member names are unique
func (r *mysqlGroupReplicationAdmission) validateUniqueNodeNames(instance *v1alpha1.MysqlGroupReplication) field.ErrorList {
	var allErrs field.ErrorList
	nodeNames := make(map[string]bool)

	// Check member names
	for i, member := range instance.Spec.Member {
		if member.Name != "" {
			if nodeNames[member.Name] {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath("spec").Child("member").Index(i).Child("name"),
					member.Name,
					"group member names must be unique",
				))
			}
			nodeNames[member.Name] = true
		}
	}

	return allErrs
}

// validateSecret validates secret configuration
func (r *mysqlGroupReplicationAdmission) validateSecret(instance *v1alpha1.MysqlGroupReplication) field.ErrorList {
	var allErrs field.ErrorList
	secretPath := field.NewPath("spec").Child("secret")

	if instance.Spec.Secret.Name == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("name"),
			"secret name is required",
		))
	}

	if instance.Spec.Secret.Mysql == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("mysql"),
			"mysql secret key is required",
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
	// DNS labels are the individual components of a domain name (e.g., "mysql", "example", "com").
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
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
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
