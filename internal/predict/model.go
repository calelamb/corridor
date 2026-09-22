// Package predict estimates spatial support of published migration routes.
// It does not estimate collision probabilities or future animal trajectories.
package predict

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math"
	"sort"
)

const Version = "pequop-kernel-v1"
const MaxCells = 2048

var ErrSupport = errors.New("insufficient released movement support")

type Sample struct {
	Cell      string          `json:"cell"`
	Block     string          `json:"block"`
	X         float64         `json:"x"`
	Y         float64         `json:"y"`
	Routes    int             `json:"routes"`
	Road      bool            `json:"road"`
	Crossings int             `json:"crossings"`
	Geometry  json.RawMessage `json:"geometry"`
}
type Model struct {
	BandwidthKM float64    `json:"bandwidth_km"`
	Evaluation  Evaluation `json:"evaluation"`
	Digest      string     `json:"input_sha256"`
	Estimates   []float64  `json:"-"`
	Samples     []Sample   `json:"-"`
}

func Fit(input []Sample) (Model, error) {
	return FitContext(context.Background(), input)
}

func FitContext(ctx context.Context, input []Sample) (Model, error) {
	if err := ctx.Err(); err != nil {
		return Model{}, err
	}
	if err := validate(input); err != nil {
		return Model{}, err
	}
	rows := append([]Sample(nil), input...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Cell < rows[j].Cell })
	h, e, err := evaluate(ctx, rows)
	if err != nil {
		return Model{}, err
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return Model{}, err
	}
	sum := sha256.Sum256(encoded)
	estimates := make([]float64, len(rows))
	for i, s := range rows {
		if err := ctx.Err(); err != nil {
			return Model{}, err
		}
		estimates[i] = math.Expm1(estimate(rows, s, h))
	}
	return Model{h, e, hex.EncodeToString(sum[:]), estimates, rows}, nil
}
func validate(rows []Sample) error {
	if len(rows) < 30 || len(rows) > MaxCells {
		return ErrSupport
	}
	seen := map[string]bool{}
	positive := 0
	for _, r := range rows {
		if r.Cell == "" || r.Block == "" || seen[r.Cell] || r.Routes < 0 || r.Routes > 10000 || r.Crossings < 0 || r.Crossings > r.Routes || !finite(r.X) || !finite(r.Y) || math.Abs(r.X) > 1e7 || math.Abs(r.Y) > 1e7 {
			return errors.New("invalid model input")
		}
		seen[r.Cell] = true
		if r.Routes > 0 {
			positive++
		}
	}
	if positive < 10 {
		return ErrSupport
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func fold(block string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(block))
	return int(h.Sum32() % 5)
}

// Distance-shifted Gaussian weights avoid underflow far from the fitted samples.
func estimate(train []Sample, target Sample, km float64) float64 {
	minD := math.Inf(1)
	for _, s := range train {
		d := distance2(s, target)
		if d < minD {
			minD = d
		}
	}
	numerator, denominator := 0.0, 0.0
	for _, s := range train {
		w := math.Exp(-(distance2(s, target) - minD) / (2 * km * km * 1e6))
		numerator += w * math.Log1p(float64(s.Routes))
		denominator += w
	}
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
func distance2(a, b Sample) float64 { return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y) }
