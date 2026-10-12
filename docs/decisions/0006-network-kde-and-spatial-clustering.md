# ADR 0006: Network Kernel Density Estimation and Getis-Ord Gi* Hotspot Detection

Date: 2026-10-12. Status: Accepted.

## Context

Wildlife-vehicle collisions (WVC) occur exclusively along linear transportation infrastructure (road networks). Conventional planar 2D Kernel Density Estimation (KDE) treats geographical space isotropically across continuous Euclidean planes ($x, y$). Applying 2D planar KDE to roadkill incidents spreads estimated collision density into adjacent off-road landscape—such as open forests, lakes, agricultural fields, and mountainous terrain where vehicles cannot travel. This planar bleed distorts risk assessments by artificially deflating linear density along highway corridors while assigning false collision risk to uninhabited wildlands.

State departments of transportation (DOTs), Federal Highway Administration (FHWA) engineers, and wildlife agencies require statistically defensible, auditable spatial clustering techniques to justify capital project allocations. These methods must operate strictly along road network topology, distinguish statistically significant collision clusters from random Poisson occurrences, and run with zero external GIS service dependencies.

## Decision

We implement Network Kernel Density Estimation (NKDE) and Getis-Ord $G_i^*$ spatial autocorrelation directly in pure Go within `internal/hotspot/`:

1. **Network Kernel Density Estimation (NKDE)**:
   - Evaluates incident density along the 1D graph topology of the road network rather than 2D Euclidean space.
   - Computes network shortest-path distances between incidents and road segment midpoints using Dijkstra's algorithm bounded by bandwidth thresholds ($h$).
   - Implements both Epanechnikov ($K(u) = \frac{3}{4h}(1 - u^2)$ for $u \le 1$) and Gaussian ($K(u) = \frac{1}{h\sqrt{2\pi}} e^{-0.5 u^2}$) network kernel weighting functions.
2. **Getis-Ord $G_i^*$ Spatial Clustering**:
   - Calculates local spatial autocorrelation statistics across road segment graph neighborhoods to identify localized concentrations of high risk (hotspots) and low risk (coldspots).
   - Measures local sum deviation against global network mean and standard deviation:
     $$G_i^* = \frac{\sum_j w_{ij} x_j - \bar{X} \sum_j w_{ij}}{S \sqrt{\frac{n \sum_j w_{ij}^2 - (\sum_j w_{ij})^2}{n - 1}}}$$
   - Supports both topological 1-hop adjacency and network distance threshold neighborhoods.
3. **Statistical Significance and Multiple Testing Correction**:
   - Implements asymptotic standard normal distribution $p$-values and Monte Carlo permutation tests (Fisher-Yates shuffling across $N$ permutations) for empirical, distribution-free significance testing.
   - Applies the Benjamini-Hochberg False Discovery Rate (FDR) procedure to correct for multiple hypothesis testing across all evaluated road segments, categorizing hotspots into 90%, 95%, and 99% confidence tiers.
4. **Pure Go Implementation**:
   - All graph representation, Dijkstra traversal, permutation loops, and FDR algorithms are implemented natively in Go without external GIS engines (PostGIS, GDAL, QGIS, or ArcGIS).

## Consequences

- **Methodological Soundness**: Hotspots strictly follow highway centerlines, conforming to state DOT standards and peer-reviewed spatial transportation safety literature (Okabe et al., Bil et al.).
- **Statistical Defensibility**: Eliminates arbitrary visual clustering by applying rigorous $G_i^*$ hypothesis testing and Benjamini-Hochberg FDR control.
- **Self-Contained Deployment**: Graph algorithms and statistical clustering execute within the Corridor Go binary without external spatial database extensions or Python analytics daemons.
- **Computational Scaling**: Bounding Dijkstra searches to the kernel bandwidth distance and caching shortest paths prevents quadratic explosion on large regional road networks.

## Alternatives

- **Planar 2D KDE**: Rejected because planar smoothing spreads collision density off the road network into inaccessible terrain, producing inaccurate spatial exposure metrics.
- **Delegating to PostGIS / pgRouting**: Rejected for core runtime clustering to ensure the analytics logic is portable, testable in isolation, and independent of external database extensions.
- **Uncorrected $p$-Values**: Rejected because evaluating thousands of highway segments simultaneously produces high false-positive discovery rates without FDR adjustment.
