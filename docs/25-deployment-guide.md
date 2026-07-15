# 25 - Deployment Guide

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the deployment architecture and operational principles of Hublio.

It does not contain environment-specific deployment commands.

---

# 2. Deployment Philosophy

Hublio is deployed as a modular monolith.

All business modules are packaged into one backend service.

Background workers run as separate processes using the same codebase.

Deployment should remain simple for Version 1.

---

# 3. Runtime Components

Version 1 consists of the following services.

Backend

* Go API

Workers

* Queue Worker
* Scheduler

Frontend

* Next.js

Infrastructure

* PostgreSQL
* Redis
* Object Storage

Reverse Proxy

* Nginx

---

# 4. Production Architecture

```text id="o4k5nv"
                    Internet
                        │
                        ▼
                    Load Balancer
                        │
                        ▼
                     Nginx
              ┌─────────┴─────────┐
              ▼                   ▼
        Go API Server       Next.js Server
              │
              ▼
         PostgreSQL
              │
              ▼
            Redis
              │
              ▼
       Background Workers
```

All components communicate over private network interfaces.

---

# 5. Service Responsibilities

## Go API

Responsible for

* REST API
* Authentication
* Authorization
* Business Intent Submission

---

## Workers

Responsible for

* Execution Processing
* Retry
* Scheduling
* Event Processing

Workers are stateless.

Multiple workers may run simultaneously.

---

## Next.js

Responsible for

* Dashboard
* Authentication UI
* Platform Configuration

Business rules remain in the backend.

---

## PostgreSQL

Stores

* Configuration Data
* Runtime Data
* Events
* Audit Logs

PostgreSQL is the primary source of truth.

---

## Redis

Responsible for

* Queue
* Cache
* Distributed Locks

Redis is never the source of truth.

---

# 6. Environment Strategy

Supported environments

* Local
* Development
* Staging
* Production

Each environment should have isolated infrastructure.

---

# 7. Configuration

Configuration follows the Twelve-Factor App methodology.

Configuration must be provided through environment variables.

Secrets must never be committed to source control.

---

# 8. Containerization

Every service should be containerized.

Recommended images

* Go API
* Worker
* Next.js

Containers should be immutable.

---

# 9. Database Migration

Database migrations must run before new application instances begin serving traffic.

Migrations should be backward compatible whenever possible.

---

# 10. Horizontal Scaling

Version 1 supports horizontal scaling for

* Go API
* Workers
* Next.js

PostgreSQL remains a single primary instance.

Redis runs as a shared service.

---

# 11. Health Checks

Every service exposes

* Liveness Check
* Readiness Check

Health endpoints should not perform expensive operations.

---

# 12. Logging

All logs are written to stdout/stderr.

Log aggregation is handled by the deployment platform.

Application code should not manage log files.

---

# 13. Observability

Production deployments should provide

* Metrics
* Structured Logs
* Distributed Tracing

Monitoring is mandatory for all production services.

---

# 14. Backup Strategy

Regular backups are required for

* PostgreSQL
* Object Storage

Redis persistence is optional because Redis is not the source of truth.

---

# 15. Security

Production deployments must enforce

* HTTPS
* TLS
* Secure Cookies
* Encrypted Secrets
* Least Privilege

Internal services should not be publicly accessible.

---

# 16. Disaster Recovery

Recovery objectives

* Restore PostgreSQL
* Restore Object Storage
* Restart stateless services

Workers can be recreated without data loss.

---

# 17. Version 1 Constraints

Version 1 intentionally excludes

* Kubernetes-specific deployment
* Service Mesh
* Multi-region deployment
* Active-Active PostgreSQL

The deployment architecture prioritizes operational simplicity.

---

# 18. Guiding Principles

Keep deployment simple.

Scale stateless services horizontally.

Protect PostgreSQL.

Treat Redis as disposable.

Automate everything possible.

The platform is maintained as a single Git repository.

Each application is built independently.

Each application produces its own container image.