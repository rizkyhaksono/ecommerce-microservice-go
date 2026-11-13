# Security Policy

## Supported versions

Security fixes land on the `main` branch. This repository does not publish
versioned releases; use the latest commit on `main`.

| Branch | Supported |
| --- | --- |
| `main` | Yes |

## Reporting a vulnerability

Report security issues privately through GitHub Security Advisories:

https://github.com/rizkyhaksono/ecommerce-microservice-go/security/advisories/new

Do not open a public issue for a suspected vulnerability.

Include:

- what is affected (service, module, workflow, or image) and the impact
- the commit you tested
- steps or a proof of concept to reproduce, if you have one

You should receive an acknowledgement within 7 days. If the report is accepted,
a fix is prepared on `main` and disclosure is coordinated with you. If it is
declined, the response explains why.

## Secrets

Do not commit `.env` files, database passwords, or JWT signing keys. Copy
`.env.example` to `.env` and set `POSTGRES_PASSWORD` before running Docker
Compose. Production deploys must inject `POSTGRES_PASSWORD` and the JWT secrets
through the environment.
