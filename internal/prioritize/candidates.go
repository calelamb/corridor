package prioritize

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CandidateSite represents a continuous highway stretch proposed for wildlife mitigation.
type CandidateSite struct {
	ID                        uuid.UUID   `json:"id"`
	Segments                  []uuid.UUID `json:"segments"`
	State                     string      `json:"state"`
	HighwayRef                string      `json:"highway_ref"`
	MilepostStart             float64     `json:"milepost_start"`
	MilepostEnd               float64     `json:"milepost_end"`
	LengthM                   float64     `json:"length_m"`
	RecommendedType           string      `json:"recommended_type"`
	EstCostUSD                float64     `json:"est_cost_usd"`
	ExpectedAvoidedCollisions float64     `json:"expected_avoided_collisions"`
	ExpectedSavingsUSD        float64     `json:"expected_savings_usd"`
	BCR                       float64     `json:"bcr"`
	P10BCR                    float64     `json:"p10_bcr"`
	P90BCR                    float64     `json:"p90_bcr"`
	Rationale                 string      `json:"rationale"`
}

// SegmentRecord represents the road segment and risk score data needed for prioritization.
type SegmentRecord struct {
	ID               uuid.UUID
	HighwayRef       string
	State            string
	MilepostStart    float64
	MilepostEnd      float64
	LengthM          float64
	Coordinates      [][]float64
	RiskScore        float64
	PercentileRank   float64
	P10              float64
	P90              float64
	SpeciesBreakdown map[string]float64
}

// PrioritizeOptions configures budget limits, minimum BCR, state filtering, and risk thresholds.
type PrioritizeOptions struct {
	Budget        float64
	MinBCR        *float64
	State         *string
	MinPercentile float64
}

// GenerateCandidates generates and ranks candidate mitigation sites within a budget constraint.
func GenerateCandidates(ctx context.Context, pool *pgxpool.Pool, budget float64) ([]CandidateSite, error) {
	return GenerateCandidatesWithOptions(ctx, pool, PrioritizeOptions{
		Budget:        budget,
		MinPercentile: 90.0,
	}, DefaultConfig())
}

// GenerateCandidatesWithOptions generates candidate mitigation sites from the database with custom options and config.
func GenerateCandidatesWithOptions(ctx context.Context, pool *pgxpool.Pool, opts PrioritizeOptions, cfg PrioritizeConfig) ([]CandidateSite, error) {
	if pool == nil {
		return nil, errors.New("database pool unavailable")
	}

	minPercentile := opts.MinPercentile
	if minPercentile <= 0 {
		minPercentile = cfg.MinPercentile
	}
	if minPercentile <= 0 {
		minPercentile = 90.0
	}

	query := `
		SELECT
			r.id,
			COALESCE(r.highway_ref, ''),
			COALESCE(r.state, ''),
			COALESCE(r.milepost_start, 0),
			COALESCE(r.milepost_end, 0),
			r.length_m,
			ST_AsGeoJSON(r.geom)::text,
			s.risk_score,
			s.percentile_rank,
			s.p10,
			s.p90,
			s.species_breakdown
		FROM road_segments r
		JOIN segment_risk_scores s ON r.id = s.segment_id
		WHERE s.percentile_rank >= $1
		  AND ($2::text IS NULL OR r.state = $2)
		ORDER BY r.highway_ref, r.milepost_start, r.id
	`

	var stateFilter *string
	if opts.State != nil && *opts.State != "" {
		stateFilter = opts.State
	}

	rows, err := pool.Query(ctx, query, minPercentile, stateFilter)
	if err != nil {
		return nil, fmt.Errorf("query high-risk segments: %w", err)
	}
	defer rows.Close()

	var records []SegmentRecord
	for rows.Next() {
		var (
			rec      SegmentRecord
			geomJSON string
			species  []byte
		)
		if err := rows.Scan(
			&rec.ID,
			&rec.HighwayRef,
			&rec.State,
			&rec.MilepostStart,
			&rec.MilepostEnd,
			&rec.LengthM,
			&geomJSON,
			&rec.RiskScore,
			&rec.PercentileRank,
			&rec.P10,
			&rec.P90,
			&species,
		); err != nil {
			return nil, fmt.Errorf("scan segment record: %w", err)
		}

		if geomJSON != "" {
			var geojson struct {
				Coordinates [][]float64 `json:"coordinates"`
			}
			if err := json.Unmarshal([]byte(geomJSON), &geojson); err == nil {
				rec.Coordinates = geojson.Coordinates
			}
		}

		if len(species) > 0 {
			_ = json.Unmarshal(species, &rec.SpeciesBreakdown)
		}

		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate segment records: %w", err)
	}

	return GenerateCandidatesFromSegmentsWithOptions(records, opts, cfg), nil
}

// GenerateCandidatesFromSegments produces candidate sites from an in-memory slice of SegmentRecord.
func GenerateCandidatesFromSegments(segments []SegmentRecord, budget float64, config PrioritizeConfig) []CandidateSite {
	return GenerateCandidatesFromSegmentsWithOptions(segments, PrioritizeOptions{
		Budget:        budget,
		MinPercentile: config.MinPercentile,
	}, config)
}

// GenerateCandidatesFromSegmentsWithOptions produces candidate sites with full filter options.
func GenerateCandidatesFromSegmentsWithOptions(segments []SegmentRecord, opts PrioritizeOptions, config PrioritizeConfig) []CandidateSite {
	if len(segments) == 0 {
		return []CandidateSite{}
	}

	minPercentile := opts.MinPercentile
	if minPercentile <= 0 {
		minPercentile = config.MinPercentile
	}
	if minPercentile <= 0 {
		minPercentile = 90.0
	}

	// 1. Filter segments by state and minPercentile
	var highRisk []SegmentRecord
	for _, seg := range segments {
		if opts.State != nil && *opts.State != "" && seg.State != *opts.State {
			continue
		}
		if seg.PercentileRank >= minPercentile {
			highRisk = append(highRisk, seg)
		}
	}

	if len(highRisk) == 0 {
		return []CandidateSite{}
	}

	// 2. Group segments by highway_ref and state
	type groupKey struct {
		state      string
		highwayRef string
	}
	groups := make(map[groupKey][]SegmentRecord)
	for _, seg := range highRisk {
		k := groupKey{state: seg.State, highwayRef: seg.HighwayRef}
		groups[k] = append(groups[k], seg)
	}

	var candidates []CandidateSite

	// 3. Cluster adjacent segments in each group
	for k, segs := range groups {
		clusters := clusterAdjacent(segs)
		for _, cluster := range clusters {
			cand := buildCandidateSite(cluster, k.state, k.highwayRef, config)
			candidates = append(candidates, cand)
		}
	}

	// 4. Filter by min_bcr if requested
	if opts.MinBCR != nil {
		minB := *opts.MinBCR
		filtered := make([]CandidateSite, 0, len(candidates))
		for _, c := range candidates {
			if c.BCR >= minB {
				filtered = append(filtered, c)
			}
		}
		candidates = filtered
	}

	// 5. Sort candidates by BCR descending
	sort.Slice(candidates, func(i, j int) bool {
		if math.Abs(candidates[i].BCR-candidates[j].BCR) > 1e-6 {
			return candidates[i].BCR > candidates[j].BCR
		}
		if math.Abs(candidates[i].ExpectedSavingsUSD-candidates[j].ExpectedSavingsUSD) > 1e-2 {
			return candidates[i].ExpectedSavingsUSD > candidates[j].ExpectedSavingsUSD
		}
		return candidates[i].ID.String() < candidates[j].ID.String()
	})

	// 6. Handle budget optimization
	if opts.Budget == 0.0 {
		return []CandidateSite{}
	}
	if opts.Budget < 0.0 {
		// Negative budget means no constraint
		return candidates
	}

	// Greedy knapsack optimization under dollar budget limit
	return knapsackGreedy(candidates, opts.Budget)
}

func areAdjacent(a, b SegmentRecord) bool {
	// If both have mileposts defined
	if a.MilepostEnd > a.MilepostStart && b.MilepostEnd > b.MilepostStart {
		if math.Abs(b.MilepostStart-a.MilepostEnd) <= 0.05 || math.Abs(a.MilepostStart-b.MilepostEnd) <= 0.05 {
			return true
		}
	}

	// If coordinates are available, check distance between endpoints
	if len(a.Coordinates) >= 2 && len(b.Coordinates) >= 2 {
		aStart := a.Coordinates[0]
		aEnd := a.Coordinates[len(a.Coordinates)-1]
		bStart := b.Coordinates[0]
		bEnd := b.Coordinates[len(b.Coordinates)-1]

		if pointsTouch(aStart, bStart) || pointsTouch(aStart, bEnd) ||
			pointsTouch(aEnd, bStart) || pointsTouch(aEnd, bEnd) {
			return true
		}
	}

	return false
}

func pointsTouch(p1, p2 []float64) bool {
	if len(p1) < 2 || len(p2) < 2 {
		return false
	}
	// Longitude and latitude degree difference (~0.0005 deg is ~50m)
	return math.Hypot(p1[0]-p2[0], p1[1]-p2[1]) <= 0.0005
}

func clusterAdjacent(segs []SegmentRecord) [][]SegmentRecord {
	n := len(segs)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return [][]SegmentRecord{segs}
	}

	// Build adjacency list
	adj := make([][]int, n)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if areAdjacent(segs[i], segs[j]) {
				adj[i] = append(adj[i], j)
				adj[j] = append(adj[j], i)
			}
		}
	}

	visited := make([]bool, n)
	var clusters [][]SegmentRecord

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		var cluster []SegmentRecord
		queue := []int{i}
		visited[i] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			cluster = append(cluster, segs[curr])

			for _, neighbor := range adj[curr] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		// Sort cluster segments by MilepostStart, or ID
		sort.Slice(cluster, func(a, b int) bool {
			if cluster[a].MilepostStart != cluster[b].MilepostStart {
				return cluster[a].MilepostStart < cluster[b].MilepostStart
			}
			return cluster[a].ID.String() < cluster[b].ID.String()
		})

		clusters = append(clusters, cluster)
	}

	return clusters
}

func buildCandidateSite(cluster []SegmentRecord, state, highwayRef string, cfg PrioritizeConfig) CandidateSite {
	segIDs := make([]uuid.UUID, len(cluster))
	var (
		totalLengthM float64
		totalRisk    float64
		totalP10     float64
		totalP90     float64
		mpStart      = math.MaxFloat64
		mpEnd        = -math.MaxFloat64
	)

	speciesWeights := make(map[string]float64)

	for i, seg := range cluster {
		segIDs[i] = seg.ID
		totalLengthM += seg.LengthM
		totalRisk += seg.RiskScore
		totalP10 += seg.P10
		totalP90 += seg.P90

		if seg.MilepostStart < mpStart {
			mpStart = seg.MilepostStart
		}
		if seg.MilepostEnd > mpEnd {
			mpEnd = seg.MilepostEnd
		}

		for sp, prob := range seg.SpeciesBreakdown {
			speciesWeights[sp] += prob * math.Max(seg.LengthM, 1.0)
		}
	}

	if mpStart == math.MaxFloat64 {
		mpStart = 0
	}
	if mpEnd == -math.MaxFloat64 {
		mpEnd = 0
	}

	// Calculate weighted average species collision cost
	avgSpeciesCost := calculateAvgSpeciesCost(speciesWeights, cfg)

	// Determine recommended structure type
	structureType := recommendStructureType(cluster, totalRisk, totalLengthM, speciesWeights, cfg)
	structure, ok := cfg.Structures[structureType]
	if !ok {
		// Fallback to default
		structure = cfg.Structures["overpass_with_fencing"]
	}

	bcrRes := CalculateBCR(totalRisk, totalP10, totalP90, totalLengthM, structure, avgSpeciesCost)

	// Deterministic UUID based on segment IDs
	var segIDStrs []string
	for _, id := range segIDs {
		segIDStrs = append(segIDStrs, id.String())
	}
	sort.Strings(segIDStrs)
	hashInput := fmt.Sprintf("%s:%s:%.2f:%.2f:%s", state, highwayRef, mpStart, mpEnd, strings.Join(segIDStrs, ","))
	siteID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(hashInput))

	rationale := fmt.Sprintf("Recommended %s for %s (MP %.1f–%.1f, %.0fm). High-risk corridor with %.2f expected annual collisions (BCR %.2f [P10: %.2f, P90: %.2f]). Avoids an estimated %.1f collisions over %d-year structure lifespan. Sourced from %s.",
		structure.Name,
		highwayRef,
		mpStart,
		mpEnd,
		totalLengthM,
		totalRisk,
		bcrRes.BCR,
		bcrRes.P10BCR,
		bcrRes.P90BCR,
		bcrRes.ExpectedAvoidedCollisions,
		structure.LifespanYears,
		structure.Citation,
	)

	return CandidateSite{
		ID:                        siteID,
		Segments:                  segIDs,
		State:                     state,
		HighwayRef:                highwayRef,
		MilepostStart:             mpStart,
		MilepostEnd:               mpEnd,
		LengthM:                   totalLengthM,
		RecommendedType:           structure.ID,
		EstCostUSD:                bcrRes.LifetimeCostUSD,
		ExpectedAvoidedCollisions: bcrRes.ExpectedAvoidedCollisions,
		ExpectedSavingsUSD:        bcrRes.ExpectedSavingsUSD,
		BCR:                       bcrRes.BCR,
		P10BCR:                    bcrRes.P10BCR,
		P90BCR:                    bcrRes.P90BCR,
		Rationale:                 rationale,
	}
}

func calculateAvgSpeciesCost(weights map[string]float64, cfg PrioritizeConfig) float64 {
	var totalWeight float64
	var weightedSum float64

	for sp, w := range weights {
		if w <= 0 {
			continue
		}
		cost, ok := cfg.SpeciesCosts[sp]
		if !ok {
			cost = cfg.DefaultSpeciesCost
		}
		weightedSum += w * cost
		totalWeight += w
	}

	if totalWeight <= 0 {
		return cfg.DefaultSpeciesCost
	}
	return weightedSum / totalWeight
}

func recommendStructureType(cluster []SegmentRecord, totalRisk, totalLengthM float64, speciesWeights map[string]float64, cfg PrioritizeConfig) string {
	if cfg.DefaultStructureType != "" {
		return cfg.DefaultStructureType
	}

	// Calculate large mammal proportion (moose, elk, pronghorn)
	var largeMammalWeight, totalWeight float64
	for sp, w := range speciesWeights {
		totalWeight += w
		if sp == "moose" || sp == "elk" || sp == "pronghorn" {
			largeMammalWeight += w
		}
	}

	largeMammalFrac := 0.0
	if totalWeight > 0 {
		largeMammalFrac = largeMammalWeight / totalWeight
	}

	// Overpass recommendation: high collision rates or large mammals in open corridors
	if totalRisk >= 2.5 || totalLengthM >= 1200.0 || (largeMammalFrac >= 0.25 && totalRisk >= 1.5) {
		return "overpass_with_fencing"
	}
	if totalRisk >= 1.0 || totalLengthM >= 600.0 {
		return "underpass_with_fencing"
	}
	if totalRisk >= 0.4 || totalLengthM >= 300.0 {
		return "fencing_with_jumpouts"
	}
	return "variable_speed_signs"
}

func knapsackGreedy(candidates []CandidateSite, budget float64) []CandidateSite {
	selected := make([]CandidateSite, 0, len(candidates))
	remBudget := budget

	for _, c := range candidates {
		if c.EstCostUSD <= remBudget {
			selected = append(selected, c)
			remBudget -= c.EstCostUSD
		}
	}

	return selected
}

// Engine wraps candidate site generation with a database connection pool and configuration.
type Engine struct {
	Pool   *pgxpool.Pool
	Config PrioritizeConfig
}

// NewEngine creates a new prioritization engine.
func NewEngine(pool *pgxpool.Pool, cfg PrioritizeConfig) *Engine {
	return &Engine{Pool: pool, Config: cfg}
}

// GenerateCandidates satisfies the prioritization provider interface.
func (e *Engine) GenerateCandidates(ctx context.Context, budget float64, minBCR *float64, state *string) ([]CandidateSite, error) {
	opts := PrioritizeOptions{
		Budget:        budget,
		MinBCR:        minBCR,
		State:         state,
		MinPercentile: e.Config.MinPercentile,
	}
	return GenerateCandidatesWithOptions(ctx, e.Pool, opts, e.Config)
}

// CandidateSite looks up a candidate site by ID.
func (e *Engine) CandidateSite(ctx context.Context, id uuid.UUID) (*CandidateSite, error) {
	candidates, err := e.GenerateCandidates(ctx, -1, nil, nil)
	if err != nil {
		return nil, err
	}
	for _, c := range candidates {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}


