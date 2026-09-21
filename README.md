# Corridor

Open source wildlife–vehicle collision intelligence for safer roads and connected habitats.

**Status: research and foundation design. The application is not implemented yet.**

Corridor's intended product is a Go service with an embedded SvelteKit map, traceable collision data, road-network hotspot analysis, evaluated risk models, and transparent mitigation planning. The full specification is in [the project brief](docs/BRIEF.md).

- [Research and prior art](docs/RESEARCH.md)
- [Data sources, restrictions, and acquisition checksums](data/SOURCES.md)
- [Phase progress and remaining dependencies](docs/PROGRESS.md)
- [Foundation design for review](docs/superpowers/specs/2026-09-20-foundation-design.md)
- [Evidence and data-release decision](docs/decisions/0001-evidence-and-data-release.md)

Code will use Apache-2.0 as specified in the brief. Datasets retain their own licenses and restrictions. Raw research downloads are excluded from Git. There is no runnable quickstart, trained model, public demo, or claim of statewide data coverage at this stage.
