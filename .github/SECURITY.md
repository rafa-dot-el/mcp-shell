# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in MCP Shell, please report it privately using our coordinated disclosure process.

### Preferred Method: Private Vulnerability Reporting

**🔒 Use GitHub's Private Vulnerability Reporting (Recommended)**

1. Go to the [Security tab](https://github.com/rafa-dot-el/mcp-shell/security)
2. Click **"Report a vulnerability"**
3. Fill out the vulnerability report form
4. Submit the report

This ensures your report is handled securely and privately until we can address it.

### Alternative Methods

1. **Email**: Send details to <security@example.com> (replace with actual email)
2. **Subject**: Include "SECURITY" and the project name in the subject line
3. **Content**: Provide a clear description of the vulnerability, including:
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

### ⚠️ Important: Do NOT use public issues for security vulnerabilities

Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 5 business days
- **Resolution Timeline**: We aim to resolve critical vulnerabilities within 30 days
- **Disclosure**: We follow responsible disclosure practices

### Security Features

This project includes several security measures:

- **Automated Security Scanning**: GitHub Advanced Security with CodeQL
- **Dependency Scanning**: Dependabot for automated dependency updates
- **Secret Scanning**: GitHub secret scanning for credential detection
- **Vulnerability Scanning**: Trivy for container and code vulnerability scanning
- **Supply Chain Security**: Go module verification and checksums

### Secure Development Practices

- All dependencies are regularly updated via Dependabot
- Security scans run on every pull request
- Code reviews are required for all changes
- Principle of least privilege in CI/CD workflows
- Regular security audits and vulnerability assessments

## Security Contact

For urgent security matters, contact:
- Email: <security@example.com>
- Response time: Within 24 hours for critical issues

## Bug Bounty

Currently, we do not offer a bug bounty program, but we appreciate and acknowledge security researchers who responsibly disclose vulnerabilities.

## Attribution

We maintain a security acknowledgments section to recognize researchers who help improve our security:

- [Security Advisories](https://github.com/rafa-dot-el/mcp-shell/security/advisories)