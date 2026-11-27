# Security review — Beauty Hub

Status: partial source review and hardening; not a penetration test or production certification. Build, integration tests and vulnerability scan results are unverified.

| Vulnerability / gap | Risk | Fix applied / remaining | Files |
|---|---|---|---|
| Unbounded DB connections | High | Pool limits, lifetime, startup ping timeout; per-query deadlines still require controller context propagation | config/database.go |
| API flooding | High | Per-IP global 180/min and auth 10/min, bounded map; distributed limiter pending | middleware/security.go, routes/routes.go |
| Oversized requests | High | 1 MiB body limit, 4 KiB query limit, request context deadline; handlers must honor context | middleware/security.go |
| Weak registration passwords | Medium | bcrypt retained, 12–72 bytes and character classes | services/auth.go |
| JWT malformed input | Medium | Length cap and issued-at validation; existing HS256 issuer and expiry checks retained; revocation pending | services/auth.go |
| Missing login audit | Medium | Structured attempt/success/failure logs without passwords or tokens | controllers/auth.go |
| Container root runtime | High | Non-root multi-stage image; image build and scan pending | Dockerfile, .dockerignore |
| Dependency exposure | Unassessed | CI govulncheck configured; no results or verified upgrades yet | .github/workflows/ci.yml |
| Cross-origin requests | High | Configured origin rejection; existing origin-based CSRF checks retained; explicit token CSRF and login CSRF review pending | middleware/security.go |
| Inline script CSP | Medium | Existing headers retained; remove inline scripts and enforce strict CSP after browser tests | middleware/security.go |
| HTTPS enforcement | High | Existing HSTS and secure cookies retained; trusted reverse proxy and HTTPS redirect configuration pending | main.go |
| SQL injection / XSS / IDOR | Unassessed | Full controller and frontend review pending; no claim of complete coverage | controllers/, assets/, views/ |
| Upload security | Unassessed | No multipart upload endpoint in inspected routes; URL-based image endpoint needs review | routes/routes.go, controllers/catalog.go |
| Password reset/change | Unassessed | No such endpoints in inspected routes; secure implementation and audit pending | routes/routes.go |
| DB privilege and secret management | High | Environment-based configuration retained; dedicated runtime role, private networking, secret rotation and history scan pending | config/database.go |

## Production release gates
- Run build, race tests, PostgreSQL integration tests, govulncheck and container scan; resolve findings.
- Audit all SQL construction, input bounds, authorization, rendering sinks and state-changing endpoints.
- Enforce HTTPS at a trusted reverse proxy; keep backend and database private. Configure APP_ORIGIN to the exact HTTPS origin, COOKIE_SECURE=true, unique random JWT_SECRET and a restricted database account. Never use demo credentials in production.
- Add session revocation, explicit CSRF protection, durable audit logs and distributed rate limiting as needed.
- Do not claim automatic UPI verification without a verified payment provider.
- No UI redesign or business workflow change was intended in this pass.
