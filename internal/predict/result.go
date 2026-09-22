package predict

import (
	"encoding/json"
	"math"
	"sort"
)

type Properties struct {
	Cell      string  `json:"cell"`
	Score     float64 `json:"score"`
	Estimate  float64 `json:"estimate"`
	Routes    int     `json:"observed_routes"`
	Crossings int     `json:"mapped_road_intersections"`
	Road      bool    `json:"road"`
	Rank      int     `json:"rank"`
}
type Feature struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties Properties      `json:"properties"`
}
type Result struct {
	Type           string    `json:"type"`
	Features       []Feature `json:"features"`
	Model          Model     `json:"model"`
	Version        string    `json:"version"`
	Study          string    `json:"study"`
	Period         string    `json:"period"`
	CandidateCount int       `json:"candidate_count"`
	SourceURL      string    `json:"source_url"`
	RoadSourceURL  string    `json:"road_source_url"`
	License        string    `json:"license"`
}

func Build(m Model) Result {
	max := 0.0
	for _, v := range m.Estimates {
		max = math.Max(max, v)
	}
	features := make([]Feature, 0, len(m.Samples))
	for i, s := range m.Samples {
		score := 0.0
		if max > 0 {
			score = 100 * m.Estimates[i] / max
		}
		features = append(features, Feature{"Feature", s.Cell, s.Geometry, Properties{s.Cell, score, m.Estimates[i], s.Routes, s.Crossings, s.Road, 0}})
	}
	candidates := []int{}
	for i, f := range features {
		if f.Properties.Road {
			candidates = append(candidates, i)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := features[candidates[i]], features[candidates[j]]
		if a.Properties.Score == b.Properties.Score {
			return a.ID < b.ID
		}
		return a.Properties.Score > b.Properties.Score
	})
	ranks := map[int]int{}
	for i, index := range candidates {
		ranks[index] = i + 1
	}
	ranked := make([]Feature, len(features))
	for i, f := range features {
		p := f.Properties
		p.Rank = ranks[i]
		ranked[i] = Feature{f.Type, f.ID, f.Geometry, p}
	}
	return Result{"FeatureCollection", ranked, m, Version, "Pequop mule deer / I-80", "2011–2017 migration study; roads snapshot 2026-05-31", len(candidates), "https://doi.org/10.5066/P9O2YM6I", "https://www.openstreetmap.org/copyright", "Migration: CC0; road-derived screening database: ODbL 1.0, © OpenStreetMap contributors"}
}
