# Security Policy

## Supported Versions

We actively maintain and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 2.0.x   | ✅ Yes             |
| 1.x.x   | ❌ No              |

## Reporting a Vulnerability

We take the security of Compose Operator seriously. If you discover a security vulnerability, please follow these steps:

### 🔒 Private Disclosure

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security issues by emailing: **security@upmio.com**

Include the following information in your report:
- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact assessment
- Any suggested fixes (if available)

### 📅 Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution Timeline**: 30-90 days (depending on complexity)

### 🛡️ What We Cover

Security issues in scope include:
- Authentication and authorization bypasses
- Remote code execution vulnerabilities
- SQL injection or command injection
- Privilege escalation
- Information disclosure
- Denial of Service (DoS) attacks

### ⚠️ Out of Scope

The following are generally considered out of scope:
- Issues in third-party dependencies (report directly to upstream)
- Social engineering attacks
- Physical security issues
- Issues requiring physical access to servers

### 🏆 Recognition

We appreciate security researchers who help keep Compose Operator secure:
- We will acknowledge your contribution in our security advisories (if desired)
- We maintain a list of security contributors in our documentation
- For significant vulnerabilities, we may offer public recognition

### 📋 Security Best Practices

For users deploying Compose Operator:
- Always use the latest supported version
- Follow the principle of least privilege for RBAC
- Use encrypted secrets for database credentials
- Enable admission webhooks for validation
- Monitor logs for suspicious activities
- Keep underlying Kubernetes cluster updated

## Security Features

Compose Operator includes several security features:
- **AES-256-CTR Encryption** for database passwords
- **RBAC Integration** with minimal required permissions
- **Admission Webhooks** for resource validation
- **TLS Support** for database connections
- **Secret Management** via Kubernetes Secrets

For more details, see our [Architecture Documentation](docs/architecture.md#security-model).