package predict

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
)

func synthetic() []Sample {
	var rows []Sample
	for i := 0; i < 160; i++ {
		rows = append(rows, Sample{Cell: fmt.Sprint(i), Block: fmt.Sprint(i / 4), X: float64(i%20) * 2000, Y: float64(i/20) * 4000, Routes: int(100 * math.Exp(-math.Pow(float64(i%20-10)/4, 2)))})
	}
	return rows
}
func TestModelEvaluation(t *testing.T) {
	rows := synthetic()
	m, err := Fit(rows)
	if err != nil {
		t.Fatal(err)
	}
	if m.Evaluation.TestCells == 0 || m.Evaluation.ModelRMSE >= m.Evaluation.BaselineRMSE {
		t.Fatalf("evaluation %+v", m.Evaluation)
	}
	changed := append([]Sample(nil), rows...)
	for i, s := range changed {
		if fold(s.Block) == 0 {
			changed[i] = Sample{Cell: s.Cell, Block: s.Block, X: s.X, Y: s.Y, Routes: 999}
		}
	}
	other, err := Fit(changed)
	if err != nil {
		t.Fatal(err)
	}
	if other.BandwidthKM != m.BandwidthKM {
		t.Fatal("test labels leaked into selection")
	}
	if rows[0].Routes != synthetic()[0].Routes {
		t.Fatal("mutated inputs")
	}
	for _, v := range m.Estimates {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			t.Fatal("invalid estimate")
		}
	}
}
func TestMissingOrInvalidData(t *testing.T) {
	for _, r := range [][]Sample{nil, {{Cell: "a", Block: "a", Routes: -1}}, {{Cell: "a", X: math.NaN()}}} {
		if _, e := Fit(r); e == nil {
			t.Fatal("accepted invalid data")
		}
	}
}

func TestCancelledFit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := FitContext(ctx, synthetic()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
