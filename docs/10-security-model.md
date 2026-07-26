# Security Model

> Product: Hublio
> Version: 1.0
> Status: Approved

---

# 1. Purpose

This document defines the security architecture of Hublio.

Security is a platform capability.

Every platform component must follow this security model.

Security is applied consistently across

- Platform API
- Orchestration Engine
- Transformation Engine
- Connector Runtime
- Event Platform
- Persistence Layer

---

# 2. Security Principles

The platform follows

Zero Trust

Least Privilege

Defense in Depth

Secure by Default

Encryption Everywhere

Immutable Audit

Security by Design

Every request is authenticated.

Every action is authorized.

Every operation is auditable.

---

# 3. Security Layers

Identity

↓

Authentication

↓

Authorization

↓

Credential Protection

↓

Execution Protection

↓

Data Protection

↓

Infrastructure Protection

Every layer is independent.

---

# 4. Identity

Identity represents

Human Users

Service Accounts

API Keys

Machine Clients

Connectors

Every identity has

Unique Identifier

Organization

Workspace

Permissions

Status

---

# 5. Authentication

Supported methods

OAuth2

JWT

API Keys

Service Accounts

Human social login (partial)

- Google / Microsoft / GitHub OAuth for **non-platform-admin** users
- Hybrid: link existing tenant user by verified email, or Organization onboarding for first login
- Platform admins use password login only (`users.is_platform_admin`)

MFA / 2FA (TOTP)

- Second factor for password login, stored per user in `user_mfa` (Identity configuration
  entity, not a Runtime Aggregate)
- The TOTP secret is encrypted at rest with AES-GCM (`MFA_ENCRYPTION_KEY`, falling back to
  `CREDENTIAL_ENCRYPTION_KEY`); recovery codes are stored only as bcrypt hashes
- Login with MFA enabled issues **no tokens**: it returns a challenge
  (`{ "mfa_required": true, "mfa_token": "..." }`) backed by a 5 minute Redis key
  `mfa_challenge:<sha256(token)>`, allowing 5 attempts
- `POST /auth/mfa/verify` accepts either a TOTP code (30s period, ±1 step of skew) or a
  single-use recovery code, then issues the normal login tokens
- `trust_device` stores `mfa_trust:<user_id>:<sha256(device_id)>` for 30 days so the challenge
  is skipped on that device
- Enrollment: `POST /auth/mfa/setup` (JWT) returns the secret, `otpauth://` URL and recovery
  codes **once**, with `enabled_at` NULL; `POST /auth/mfa/enable` confirms a code and activates
  it; `POST /auth/mfa/disable` requires the account password
- Seeded demo accounts ship without MFA — enable it per account through the setup + enable
  endpoints
- OAuth-only accounts cannot enrol: MFA guards password login

Future

OIDC (enterprise IdP)

SAML

Passkeys

Authentication proves identity.

Nothing more.

---

# 6. Authorization

Authorization determines

What can be accessed.

Permissions are evaluated against

Organization

Workspace

Connection

Capability

Resource

Operation

Authorization must never depend on connector implementation.

---

# 7. Multi-tenancy

Every request belongs to

One Organization

One Workspace

No execution may cross tenant boundaries.

Tenant isolation is mandatory.

---

# 8. API Security

Every API request requires

Authentication

Authorization

Rate Limiting

Audit

Correlation ID

Sensitive endpoints require

Additional verification when appropriate.

---

# 9. API Keys

API Keys represent machine identities.

API Keys belong to a Workspace.

Hierarchy

Organization

↓

Workspace

↓

API Key

API Keys should support

Expiration

Rotation

Revocation

Usage tracking

---

# 10. Service Accounts

Long-running integrations should use

Service Accounts

instead of

Human Users.

Service Accounts have

Minimal permissions.

---

# 11. Credential Protection

Credentials include

OAuth Tokens

Refresh Tokens

API Keys

Certificates

Private Keys

Secrets

Rules

Encrypted at rest

Encrypted in transit

Never logged

Never returned through APIs

Rotation supported

---

# 12. Secret Management

Secrets should never exist

inside source code

or

configuration files.

Production deployments should support

External Secret Management.

Examples

Vault

Cloud Secret Manager

Kubernetes Secrets

---

# 13. Connector Security

Connectors never own credentials.

Connectors receive credentials

through secure runtime context.

Connectors must never

Persist credentials

Log credentials

Expose credentials

---

# 14. Execution Security

Every execution records

Who initiated it

When

Why

Which connection

Which permissions

Execution context is immutable.

---

# 15. Event Security

Events should never expose

Passwords

Secrets

Private Keys

Access Tokens

Sensitive fields should support masking.

---

# 16. Snapshot Security

Execution snapshots may contain

Business data.

Snapshots should support

Encryption

Access control

Retention

Sensitive field masking

---

# 17. Data Classification

Data should be classified.

Public

Internal

Confidential

Restricted

Security policy depends on classification.

---

# 18. Encryption

Encryption in Transit

TLS

Encryption at Rest

Database Encryption

Object Storage Encryption

Credential Encryption

Sensitive payloads should remain encrypted whenever practical.

---

# 19. Audit Trail

Every security-sensitive action is audited.

Examples

Login

Logout

Permission Change

Credential Update

API Key Creation

Execution Replay

Connector Installation

Audit records are immutable.

---

# 20. Session Security

User sessions support

Expiration

Revocation

Device Tracking

Concurrent Session Management

MFA (TOTP + recovery codes, see §5)

Future

Passkeys

---

# 21. Rate Limiting

Limits may apply per

Organization

Workspace

User

API Key

Connection

Resource

Capability

---

# 22. Input Validation

Every external input

must be validated.

Examples

REST API

Webhook

Connector Payload

Configuration

Validation occurs

before business processing.

---

# 23. Output Protection

Sensitive fields should never be exposed unnecessarily.

Examples

Secrets

Credentials

Internal IDs

Private Metadata

Masking should be configurable.

---

# 24. Logging

Logs should never contain

Passwords

Secrets

Access Tokens

Private Keys

PII should support masking.

---

# 25. Replay Protection

Idempotency prevents

Duplicate business operations.

Replay protection should apply to

REST

Webhook

Connector callbacks

---

# 26. Platform Isolation

Every platform component

is isolated.

API

↓

Orchestration

↓

Transformation

↓

Connector Runtime

↓

Persistence

No layer bypasses security checks.

---

# 27. Dependency Security

Third-party dependencies should

Be tracked

Be scanned

Be updated

Known vulnerabilities should be monitored continuously.

---

# 28. Infrastructure Security

Production environments should support

Network Segmentation

Firewall

Private Networking

TLS Everywhere

Least Privilege IAM

Infrastructure security remains independent from application logic.

---

# 29. Incident Response

Security incidents should support

Detection

Alerting

Investigation

Containment

Recovery

Audit history should assist forensic analysis.

---

# 30. Security Principles

Security belongs to the platform.

Connectors inherit platform security.

Every capability should be

Authenticated

Authorized

Audited

Observable

Recoverable

Security is not a feature.

Security is a platform capability.