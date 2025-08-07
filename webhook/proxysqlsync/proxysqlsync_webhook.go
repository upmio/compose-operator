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
var proxysqlsynclog = ctrl.Log.WithName("proxysql-sync").WithValues("version", "v1alpha1")

type proxysqlSyncAdmission struct {
}

// Setup will setup the manager to manage the webhooks
func Setup(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.ProxysqlSync{}).
		WithValidator(&proxysqlSyncAdmission{}).
		WithDefaulter(&proxysqlSyncAdmission{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-upm-syntropycloud-io-v1alpha1-proxysqlsync,mutating=true,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=proxysqlsyncs,verbs=create;update,versions=v1alpha1,name=mproxysqlsync.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &proxysqlSyncAdmission{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *proxysqlSyncAdmission) Default(ctx context.Context, obj runtime.Object) error {
	instance := obj.(*v1alpha1.ProxysqlSync)
	proxysqlsynclog.Info("default", "name", instance.Name)

	// Set default secret key names
	if instance.Spec.Secret.Proxysql == "" {
		instance.Spec.Secret.Proxysql = "proxysql"
	}
	if instance.Spec.Secret.Mysql == "" {
		instance.Spec.Secret.Mysql = "mysql"
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-upm-syntropycloud-io-v1alpha1-proxysqlsync,mutating=false,failurePolicy=fail,sideEffects=None,groups=upm.syntropycloud.io,resources=proxysqlsyncs,verbs=create;update,versions=v1alpha1,name=vproxysqlsync.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &proxysqlSyncAdmission{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *proxysqlSyncAdmission) ValidateCreate(ctx context.Context, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.ProxysqlSync)
	proxysqlsynclog.Info("validate create", "name", instance.Name)

	return r.validateProxysqlSync(instance)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *proxysqlSyncAdmission) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	instance := newObj.(*v1alpha1.ProxysqlSync)

	proxysqlsynclog.Info("validate update", "name", instance.Name)

	oldPS := oldObj.(*v1alpha1.ProxysqlSync)

	// Validate the new object
	warnings, err = r.validateProxysqlSync(instance)
	if err != nil {
		return warnings, err
	}

	// Additional update-specific validations
	var allErrs field.ErrorList

	// Secret reference should not change
	if oldPS.Spec.Secret.Name != instance.Spec.Secret.Name {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec").Child("secret").Child("name"),
			instance.Spec.Secret.Name,
			"secret reference cannot be changed after creation",
		))
	}

	// MysqlReplication reference should not change
	if oldPS.Spec.MysqlReplication != instance.Spec.MysqlReplication {
		warnings = append(warnings, "Changing MysqlReplication reference may cause synchronization issues.")
	}

	if len(allErrs) > 0 {
		return warnings, allErrs.ToAggregate()
	}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *proxysqlSyncAdmission) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	instance := obj.(*v1alpha1.ProxysqlSync)

	proxysqlsynclog.Info("validate delete", "name", instance.Name)

	// Add warning about potential impact
	warnings = append(warnings, "Deleting ProxysqlSync will stop managing ProxySQL synchronization but won't affect existing ProxySQL instances.")

	return warnings, nil
}

// validateProxysqlSync performs comprehensive validation of ProxysqlSync spec
func (r *proxysqlSyncAdmission) validateProxysqlSync(instance *v1alpha1.ProxysqlSync) (admission.Warnings, error) {
	var allErrs field.ErrorList
	var warnings admission.Warnings

	// Validate ProxySQL nodes
	if len(instance.Spec.Proxysql) == 0 {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec").Child("proxysql"),
			"at least one ProxySQL node is required",
		))
	} else {
		for i, proxysql := range instance.Spec.Proxysql {
			fieldPath := field.NewPath("spec").Child("proxysql").Index(i)
			if err := r.validateCommonNode("proxysql", proxysql, fieldPath); err != nil {
				allErrs = append(allErrs, err...)
			}
		}

		// Check for duplicate ProxySQL node names
		if err := r.validateUniqueNodeNames(instance); err != nil {
			allErrs = append(allErrs, err...)
		}
	}

	// Validate secret configuration
	if err := r.validateSecret(instance); err != nil {
		allErrs = append(allErrs, err...)
	}

	// Validate MysqlReplication reference
	if instance.Spec.MysqlReplication == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec").Child("mysqlReplication"),
			"MysqlReplication reference is required",
		))
	}

	if len(allErrs) > 0 {
		return warnings, allErrs.ToAggregate()
	}

	return warnings, nil
}

// validateCommonNode validates common node fields
func (r *proxysqlSyncAdmission) validateCommonNode(nodeType string, node *v1alpha1.CommonNode, fieldPath *field.Path) field.ErrorList {
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

// validateUniqueNodeNames ensures all ProxySQL node names are unique
func (r *proxysqlSyncAdmission) validateUniqueNodeNames(instance *v1alpha1.ProxysqlSync) field.ErrorList {
	var allErrs field.ErrorList
	nodeNames := make(map[string]bool)

	// Check ProxySQL node names
	for i, proxysql := range instance.Spec.Proxysql {
		if proxysql.Name != "" {
			if nodeNames[proxysql.Name] {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath("spec").Child("proxysql").Index(i).Child("name"),
					proxysql.Name,
					"ProxySQL node names must be unique",
				))
			}
			nodeNames[proxysql.Name] = true
		}
	}

	return allErrs
}

// validateSecret validates secret configuration
func (r *proxysqlSyncAdmission) validateSecret(instance *v1alpha1.ProxysqlSync) field.ErrorList {
	var allErrs field.ErrorList
	secretPath := field.NewPath("spec").Child("secret")

	if instance.Spec.Secret.Name == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("name"),
			"secret name is required",
		))
	}

	if instance.Spec.Secret.Proxysql == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("proxysql"),
			"proxysql secret key is required",
		))
	}

	if instance.Spec.Secret.Mysql == "" {
		allErrs = append(allErrs, field.Required(
			secretPath.Child("mysql"),
			"mysql secret key is required",
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
	// DNS labels are the individual components of a domain name (e.g., "proxysql", "example", "com").
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
