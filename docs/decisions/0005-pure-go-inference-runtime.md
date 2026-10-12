# ADR 0005: Pure Go Model Inference Runtime for LightGBM and GLM Models

Date: 2026-10-12. Status: Accepted.

## Context

Production services in Corridor must serve LightGBM Gradient Boosted Decision Tree (GBDT) risk models and two-stage Hurdle Generalized Linear Models (GLM) for real-time segment scoring, route risk calculations, and batch spatial prioritization. The architecture mandates that all production backend services compile as static binaries with zero CGo dependencies (`CGO_ENABLED=0`), ensuring fast, portable compilation, small container images, and no external runtime C/C++ library dependencies.

The Corridor engineering contract dictates that Python is restricted to offline training, feature selection, and model validation in `ml/`. Production inference engines written in Python or wrapped via CGo (such as linking against `lib_lightgbm.so` or `libgomp`) introduce substantial deployment friction: complex toolchains, libc compatibility hazards (e.g., glibc vs. musl on Alpine), elevated container vulnerabilities, and garbage collection / memory management overhead across language boundaries.

## Decision

We implement a pure Go model inference runtime in `internal/model/` that evaluates trained LightGBM trees and Hurdle GLM models directly:

1. **Pure Go LightGBM Tree Evaluator**: Load serialized LightGBM text models using the pure Go `leaves` parser (`github.com/dmitryikh/leaves`). To support modern LightGBM exports (v4+), the Go loader normalizes version header tokens to `version=v3` on the fly, matching the grammar supported by `leaves` while retaining complete tree topology and feature threshold definitions.
2. **Hurdle GLM Runtime**: Load two-stage Hurdle Negative Binomial GLM models from JSON representations exported by `corridor_ml.models.export`. Inference executes in pure Go vector arithmetic:
   - **Stage 1**: Logistic regression hurdle probability $P(Y > 0) = \frac{1}{1 + e^{-(\beta_0 + X \beta)}}$, modeling reporting and detection probability.
   - **Stage 2**: Zero-truncated Negative Binomial count expectation $E(Y \mid Y > 0) = \frac{\mu}{1 - (1 + \alpha \mu)^{-1/\alpha}}$ where $\mu = \exp(\gamma_0 + X \gamma + \text{offset})$, incorporating logarithmic exposure offsets for road segment lengths.
   - **Expected Rate**: Overall expected collision frequency $E(Y) = P(Y > 0) \times E(Y \mid Y > 0)$.
3. **Common Predictor Interface**: Expose a unified `Predictor` interface (`Predict(features []float64) (float64, error)`) across all model families, standardizing model evaluation across the REST API, tile generators, and priority pipelines.

## Consequences

- **High Throughput & Sub-Millisecond Latency**: Tree traversal and vector arithmetic execute in microseconds per segment with minimal heap allocations.
- **Single Static Binary**: Services compile with `CGO_ENABLED=0`, producing fully self-contained Go binaries deployable to scratch or minimal Alpine containers.
- **Strict Boundary Separation**: Python data science dependencies (LightGBM, scikit-learn, statsmodels, NumPy) remain isolated in `ml/` for offline training, keeping production images clean and lean.
- **Verification Requirement**: Deserialization parity and numerical equivalence tests must verify that pure Go evaluations match Python reference predictions within floating-point tolerance ($10^{-6}$).

## Alternatives

- **Embedding CGo / lib_lightgbm**: Rejected due to platform-specific dynamic linking issues, container bloat, and violation of the `CGO_ENABLED=0` architecture rule.
- **Python Sidecar / Microservice for Inference**: Rejected due to network serialization overhead, added service orchestration complexity, and increased cold-start latency.
- **ONNX Runtime (Go ONNX bindings)**: Rejected because official ONNX Go bindings rely on CGo and external C shared libraries (`libonnxruntime.so`).
