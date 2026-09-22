package predict

import (
	"context"
	"math"
)

type Evaluation struct {
	ModelRMSE        float64 `json:"model_rmse"`
	BaselineRMSE     float64 `json:"baseline_rmse"`
	Improvement      float64 `json:"improvement_percent"`
	TestCells        int     `json:"test_cells"`
	TestBlocks       int     `json:"test_blocks"`
	DevelopmentCells int     `json:"development_cells"`
	Passed           bool    `json:"passed"`
}

func evaluate(ctx context.Context, rows []Sample) (float64, Evaluation, error) {
	dev, test := partition(rows, 0)
	blocks := map[string]bool{}
	for _, s := range test {
		blocks[s.Block] = true
	}
	if len(dev) < 20 || len(test) < 5 || len(blocks) < 2 {
		return 0, Evaluation{}, ErrSupport
	}
	best, loss := 0.0, math.Inf(1)
	for _, h := range []float64{3, 6, 12} {
		sum, n := 0.0, 0
		for f := 1; f < 5; f++ {
			train, validation := partition(dev, f)
			if len(train) < 10 || len(validation) < 2 {
				return 0, Evaluation{}, ErrSupport
			}
			for _, s := range validation {
				if err := ctx.Err(); err != nil {
					return 0, Evaluation{}, err
				}
				delta := estimate(train, s, h) - math.Log1p(float64(s.Routes))
				sum += delta * delta
				n++
			}
		}
		if score := sum / float64(n); score < loss {
			best, loss = h, score
		}
	}
	model, baseline := errorsFor(dev, test, best)
	improvement := 0.0
	if baseline > 0 {
		improvement = 100 * (baseline - model) / baseline
	}
	return best, Evaluation{model, baseline, improvement, len(test), len(blocks), len(dev), model < baseline}, nil
}
func partition(rows []Sample, heldout int) ([]Sample, []Sample) {
	train, test := []Sample{}, []Sample{}
	for _, s := range rows {
		if fold(s.Block) == heldout {
			test = append(test, s)
		} else {
			train = append(train, s)
		}
	}
	return train, test
}
func errorsFor(train, test []Sample, h float64) (float64, float64) {
	mean := 0.0
	for _, s := range train {
		mean += math.Log1p(float64(s.Routes))
	}
	mean /= float64(len(train))
	model, baseline := 0.0, 0.0
	for _, s := range test {
		actual := math.Log1p(float64(s.Routes))
		d := estimate(train, s, h) - actual
		model += d * d
		b := mean - actual
		baseline += b * b
	}
	return math.Sqrt(model / float64(len(test))), math.Sqrt(baseline / float64(len(test)))
}
