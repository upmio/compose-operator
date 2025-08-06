---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: ['bug', 'needs-triage']
assignees: ''
---

## Bug Description

**Clear and concise description of what the bug is:**

## Steps to Reproduce

**Provide detailed steps to reproduce the behavior:**

1. Deploy operator with configuration:
   ```yaml
   # Include relevant configuration here
   ```

2. Create/apply the following custom resource:
   ```yaml
   # Include the CRD YAML here
   ```

3. Observe the following behavior:
   ```
   # Describe what you see
   ```

4. Expected behavior:
   ```
   # Describe what you expected to happen
   ```

## Environment Information

**Please complete the following information:**

- **Kubernetes Version**: [e.g., v1.28.2]
- **Operator Version**: [e.g., v2.0.1]
- **Installation Method**: [e.g., Helm, manual manifests]
- **Database Versions**: 
  - MySQL: [e.g., 8.0.34]
  - Redis: [e.g., 7.0.12]
  - PostgreSQL: [e.g., 15.4]
  - ProxySQL: [e.g., 2.5.4]
- **Cloud Provider/Platform**: [e.g., AWS EKS, GKE, on-premises]
- **Network Setup**: [e.g., CNI plugin, network policies]

## Logs and Error Messages

**Operator logs:**
```
# Paste relevant operator logs here
# Get logs with: kubectl logs -n upm-system -l app.kubernetes.io/name=compose-operator --tail=100
```

**Database logs (if relevant):**
```
# Paste relevant database logs here
```

**Kubernetes events:**
```
# Get events with: kubectl get events --field-selector involvedObject.name=<resource-name>
```

## Configuration Files

**Custom Resource Configuration:**
```yaml
# Paste the CRD configuration that's causing issues
# REMOVE SENSITIVE INFORMATION like passwords, hostnames, etc.
```

**Operator Configuration:**
```yaml
# If using custom operator configuration, paste here
# REMOVE SENSITIVE INFORMATION
```

## Additional Context

**Screenshots (if applicable):**
<!-- Drag and drop images here -->

**Related Issues:**
<!-- Link to any related issues -->

**Workarounds:**
<!-- If you found any workarounds, describe them here -->

**Impact Assessment:**
- [ ] This is blocking critical functionality
- [ ] This affects multiple users/environments
- [ ] This is a minor inconvenience
- [ ] This only affects edge cases

## Checklist

Before submitting this bug report:

- [ ] I have searched existing issues to avoid duplicates
- [ ] I have included all relevant information above
- [ ] I have removed sensitive information from configurations
- [ ] I have tested with the latest version of the operator
- [ ] I have included relevant logs and error messages

## Security Consideration

- [ ] This bug has security implications (if yes, please follow our [Security Policy](../../SECURITY.md) instead)

---

**Additional Notes:**
<!-- Any additional information that might be helpful -->