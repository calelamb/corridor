# ADR 0004: Private S3 storage and archived development companion

Accepted 2026-09-21. The MinIO repository is archived. Its final release
RELEASE.2025-10-15T17-29-55Z is built from a SHA-256-verified upstream tarball
in deploy/minio.Dockerfile because the corresponding Quay image is unavailable.
MinIO is AGPL-3.0; its license is retained in the image and source URL in the
Dockerfile. It remains a separate private development companion, not an
assertion of maintained production infrastructure. Production must select a
maintained S3-compatible service and review its operational requirements.

The API uses minio-go/v7 7.3.0 (Apache-2.0) against S3-compatible operations.
TLS certificate verification remains enabled. Only the setup CLI uses the
madmin-go/v3 3.0.110 administration SDK (Apache-2.0). Root credentials are never
passed to the HTTP service. The HTTP user can check the configured bucket but
cannot read or write original objects. The setup job ensures no anonymous
bucket policy and is safe to repeat. Ingestion credentials are future work.
