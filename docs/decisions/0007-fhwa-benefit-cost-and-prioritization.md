# ADR 0007: FHWA Wildlife Crossings Pilot Program Benefit-Cost Prioritization and PDF Export

Date: 2026-10-12. Status: Accepted.

## Context

The Federal Highway Administration (FHWA) administers the Wildlife Crossings Pilot Program (WCPP), a competitive discretionary grant program allocating $350 million under the Infrastructure Investment and Jobs Act (IIJA) to reduce wildlife-vehicle collisions (WVC) and improve terrestrial habitat connectivity. To secure federal grant awards, state departments of transportation, regional planning organizations, and tribal authorities must provide rigorous, quantitative Benefit-Cost Analyses (BCA) demonstrating economic viability (Benefit-Cost Ratio, BCR $\ge 1.0$), empirical collision reduction projections, and defensible engineering cost estimates.

Transportation decision-makers previously relied on fragmented spreadsheets, ad-hoc literature lookups, and manual report authoring. Corridor requires an end-to-end, automated prioritization engine that synthesizes modeled collision risks, structures mitigation alternatives, calculates life-cycle financial returns, and generates publication-grade grant proposals directly from the platform.

## Decision

We implement an economic prioritization engine in `internal/prioritize/` and a publication-grade PDF generator in `internal/export/`:

1. **Candidate Site Aggregation**:
   - Aggregate contiguous road segments exceeding high-risk thresholds (e.g., 90th percentile risk score) along shared highway corridors into cohesive candidate mitigation projects.
   - Enforce minimum and maximum project length constraints to reflect realistic civil engineering constructability.
2. **Peer-Reviewed Mitigation Treatment Effectiveness**:
   - Model structural countermeasures using empirical efficacy rates established in peer-reviewed wildlife ecology literature:
     - **Wildlife Overpass with Continuous Exclusion Fencing**: 86% collision reduction (Rytwinski et al., 2016 meta-analysis).
     - **Wildlife Underpass with Continuous Exclusion Fencing**: 83% collision reduction (Clevenger & Huijser, 2011).
     - **Exclusion Fencing with Escape Jumpouts**: 78% collision reduction (Huijser et al., 2009).
     - **Dynamic / Variable Wildlife Warning Signs**: 35% collision reduction (Huijser et al., 2009).
3. **Societal Crash Cost Valuations**:
   - Apply USDOT comprehensive crash valuations and wildlife collision economic models to calculate monetized societal savings:
     - **Deer**: $14,000 per collision.
     - **Pronghorn**: $16,000 per collision.
     - **Elk**: $45,000 per collision.
     - **Moose**: $83,000 per collision.
   - Economic values incorporate property damage, human injury treatment, lost economic productivity, emergency medical services, and carcass removal/disposal costs.
4. **Life-Cycle Benefit-Cost Analysis (BCA) and Portfolio Optimization**:
   - Life-Cycle Cost = Capital Construction Cost (structure base + fencing length $\times$ cost/m) + (Annual Maintenance $\times$ Lifespan).
   - Lifetime Benefit = Annual Avoided Collisions $\times$ Lifespan $\times$ Average Species Cost.
   - $\text{BCR} = \frac{\text{Lifetime Benefit}}{\text{Lifetime Cost}}$, calculated with P10 and P90 uncertainty bounds propagated from model rate percentiles.
   - Apply a greedy knapsack budget allocation algorithm to identify the optimal set of mitigation projects that maximize aggregate avoided collisions and net financial return under a specified funding ceiling.
5. **Pure Go Publication-Grade PDF Grant Briefs**:
   - Implement PDF generation in `internal/export/pdf.go` using `github.com/go-pdf/fpdf` (`CGO_ENABLED=0`).
   - Format standardized 2-page WCPP grant briefs containing:
     - **Page 1**: Executive project summary, route metadata (milepost limits, state, length), proposed mitigation structure, and geospatial corridor map snapshot.
     - **Page 2**: Key performance indicators (BCR, avoided collisions, capital vs. maintenance costs, net savings), 20-year economic life-cycle cash flow table, literature citations (Rytwinski, Clevenger, Huijser), and agency submission signature blocks.

## Consequences

- **Direct Grant Readiness**: Planners can export submission-ready 2-page WCPP funding briefs directly from the Corridor web application without external GIS or desktop publishing software.
- **Auditable Economic Prioritization**: Project selection is driven by defensible, peer-reviewed engineering costs, USDOT economic metrics, and uncertainty intervals.
- **Zero Headless Browser Dependencies**: Using `fpdf` in pure Go avoids running resource-heavy headless Chromium (Puppeteer/Playwright) or external CLI utilities (`wkhtmltopdf`/`pandoc`), keeping binary size compact and PDF generation latency under 50ms.
- **Configurable Parameters**: Mitigation cost tables, lifespan assumptions, and species valuation models are configurable via YAML to adapt to state-specific DOT rates and future FHWA guidance updates.

## Alternatives

- **Manual Spreadsheet Export (CSV only)**: Rejected because transportation agencies require standardized, visual grant documents ready for administrative review.
- **Headless Chrome PDF Rendering**: Rejected due to high memory footprint (hundreds of MBs per renderer process), slower generation times, and complex container sandboxing requirements.
- **Unconstrained Segment-by-Segment Treatment**: Rejected because wildlife crossing infrastructure requires continuous highway fencing stretches (several kilometers) to prevent animal fence-end run-arounds.
