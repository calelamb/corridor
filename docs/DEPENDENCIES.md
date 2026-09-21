# Reproducible dependencies

Go 1.27.1 is selected by go.mod; Go's toolchain resolver installs it locally.
Go tool dependencies and runtime dependencies are pinned by go.mod/go.sum.
The API generator is oapi-codegen 2.8.0 (Apache-2.0), runtime 1.6.0
(Apache-2.0), chi 5.3.2 (MIT), kin-openapi 0.142.0 (MIT).

The frontend was created using official `sv` 0.17.1 minimal TypeScript scaffold,
with static adapter, Tailwind, ESLint, Prettier, Vitest and Playwright addons.
Direct versions are exact in package.json and transitive versions/integrity
hashes live in package-lock.json. Use npm ci. Node 24.21.0 is the selected LTS
builder; the host Node installation need not change. Svelte and SvelteKit are
MIT, TypeScript Apache-2.0, Tailwind MIT, Vitest MIT, Playwright Apache-2.0.
Image digests and additional runtime dependencies are recorded with deployment.

Generated API code is excluded from authored-code size and coverage limits;
never hand-edit it. Run make generate; make check-generated requires generated
files to be tracked and free of staged or unstaged drift. No OpenAPI version
conversion was needed for this concrete contract.
