# Security policy

Corridor is an experimental research application. Only the latest `main` branch
receives fixes; there is no supported production release or response-time SLA.
The default stack binds the application to localhost. Do not expose the pilot
to the internet without a deployment security review and maintained storage.

## Report privately

Use [GitHub private vulnerability reporting](https://github.com/calelamb/corridor/security/advisories/new)
for security vulnerabilities, credential exposure, and wildlife-location privacy
failures. Include affected commit, impact, and minimal reproduction steps using
synthetic records. Do not put credentials, exact sensitive wildlife coordinates,
private observations, photos, or exploit payloads in public issues or pull requests.

Please allow the maintainer to investigate and coordinate disclosure before
publishing details. Public issues are appropriate for ordinary bugs using
non-sensitive examples; see the issue templates.

## Boundaries

- Original observations and credentials remain in ignored local files/private storage.
- The public service uses read-only database credentials and approved generalized views.
- Sensitive observations require H3 resolution 6 or coarser and a 30-day delay.
- Do not disable TLS verification, broaden CSP, or add public raw-data routes.
- Dataset access and code licensing are separate. A public download is not reuse permission.
- MinIO is an archived development companion; choose maintained storage for deployment.

Run `make security` and a history scan (`gitleaks git --redact --log-opts=--all`)
before publishing changes. The narrowly scoped `.gitleaksignore` entry documents
one verified empty-template false positive, not an excluded secret-bearing file.
