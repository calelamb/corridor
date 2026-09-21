# ADR 0003: Static CSP hashes

Accepted 2026-09-21. Prerendered SvelteKit pages use build-generated script hashes
instead of per-request nonces. Their bytes are fixed and embedded into Go.
This is a narrow exception to the general nonce-based CSP rule: regenerating
assets regenerates the matching hashes. Script policy never uses unsafe-inline.

Scripts, fonts, connections and styles come from the same origin. MapLibre's
bundled worker needs blob: in worker-src. Its layout updates require inline
styles; style-src permits unsafe-inline only for styling. No third-party tile,
font, analytics or script origin is authorized. The Go response adds frame,
MIME, referrer and permissions protections; the prerendered CSP meta element
carries the specific script hashes. HSTS is emitted only for direct HTTPS.
