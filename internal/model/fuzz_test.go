package model_test

import (
	"bytes"
	"math"
	"os"
	"testing"

	"corridor/internal/model"
)

// FuzzLoadLeavesModel exercises loading LightGBM gradient boosted tree ensembles
// from arbitrary text, binary, and corrupt inputs. It asserts the loader returns an
// error or valid Predictor, never panics, and never deadlocks.
func FuzzLoadLeavesModel(f *testing.F) {
	// Seed with valid LightGBM model from testdata
	validModel, err := os.ReadFile("testdata/lightgbm_poisson.txt")
	if err == nil {
		f.Add(validModel)
	}

	// Seed with boundary, corrupt, and binary inputs
	f.Add([]byte(""))
	f.Add([]byte("   \n\t  "))
	f.Add([]byte("not a lightgbm model"))
	f.Add([]byte("version=v4\n"))
	f.Add([]byte("version=v5\n"))
	f.Add([]byte("version=v3\nnum_class=1\nnum_tree_per_iteration=1\nlabel_index=0\nmax_feature_idx=10\n"))
	f.Add([]byte("Tree=0\nnum_leaves=2\nnum_cat=0\nsplit_feature=0\nsplit_gain=0\nthreshold=0\n"))
	f.Add([]byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd})
	f.Add([]byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01}) // ELF binary
	f.Add([]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00})       // gzip header
	f.Add([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10})       // JPEG header
	f.Add(bytes.Repeat([]byte("A"), 5000))
	f.Add(bytes.Repeat([]byte("\x00"), 1024))
	f.Add(bytes.Repeat([]byte("\n"), 2000))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip("skipping excessively large payload (>1MB)")
		}

		p, err := model.LoadLeavesModel(bytes.NewReader(data))
		if err != nil {
			if p != nil {
				t.Fatalf("expected nil predictor on error, got non-nil: %v", p)
			}
			return
		}

		// When successfully loaded, verify predictor interface invariants
		if p == nil {
			t.Fatalf("expected non-nil predictor when err is nil")
		}

		lm, ok := p.(*model.LeavesModel)
		if !ok {
			t.Fatalf("expected *model.LeavesModel, got %T", p)
		}

		nFeat := lm.NFeatures()
		if nFeat < 0 {
			t.Fatalf("invalid negative feature count: %d", nFeat)
		}

		// Exercise Predict to ensure no panic on inference
		_, _ = lm.Predict(make([]float64, nFeat))
		_, _ = lm.Predict([]float64{}) // insufficient features test
	})
}

// FuzzLoadHurdleModel exercises loading Hurdle Negative Binomial GLM models from
// arbitrary JSON, malformed text, and random binary inputs. It asserts the loader
// never panics.
func FuzzLoadHurdleModel(f *testing.F) {
	// Seed with valid hurdle model JSON from testdata
	validHurdle, err := os.ReadFile("testdata/hurdle_model.json")
	if err == nil {
		f.Add(validHurdle)
	}

	// Seed with boundary, corrupt, and binary inputs
	f.Add([]byte(""))
	f.Add([]byte("   \n\t  "))
	f.Add([]byte("{not json}"))
	f.Add([]byte(`{"model_type": "random_forest"}`))
	f.Add([]byte(`{"model_type": "hurdle_negative_binomial"}`))
	f.Add([]byte(`{"model_type": "hurdle_negative_binomial", "feature_names": []}`))
	f.Add([]byte(`{"model_type": "hurdle_negative_binomial", "feature_names": ["f1"]}`))
	f.Add([]byte(`{"model_type": "hurdle_negative_binomial", "feature_names": ["f1"], "scaler": {"means": {"f1": 0}, "stds": {"f1": 1}}, "hurdle_stage": {"intercept": 0, "coefficients": {"f1": 0.5}}, "count_stage": {"intercept": 0, "dispersion": 1.0, "coefficients": {"f1": 0.5}}, "exposure": {"default_length_m": 160.9, "reference_length_m": 160.9, "log_offset": true}}`))
	f.Add([]byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd})
	f.Add([]byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01})
	f.Add([]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00})
	f.Add(bytes.Repeat([]byte("{"), 500))
	f.Add(bytes.Repeat([]byte("0"), 5000))
	f.Add(bytes.Repeat([]byte("\x00"), 1024))
	f.Add([]byte(`{"model_type": "hurdle_negative_binomial", "feature_names": ["a", "b"], "scaler": null}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip("skipping excessively large payload (>1MB)")
		}

		p, err := model.LoadHurdleModel(bytes.NewReader(data))
		if err != nil {
			if p != nil {
				t.Fatalf("expected nil predictor on error, got non-nil: %v", p)
			}
			return
		}

		if p == nil {
			t.Fatalf("expected non-nil predictor when err is nil")
		}

		hurdle, ok := p.(*model.HurdleModel)
		if !ok {
			t.Fatalf("expected *model.HurdleModel, got %T", p)
		}

		nFeat := hurdle.NFeatures()
		if nFeat < 0 {
			t.Fatalf("invalid negative feature count: %d", nFeat)
		}

		// Exercise Predict and PredictWithLength invariants
		_, _ = hurdle.Predict(make([]float64, nFeat))
		_, _ = hurdle.Predict(make([]float64, nFeat+1))
		_, _ = hurdle.Predict([]float64{})
		_, _ = hurdle.PredictWithLength(make([]float64, nFeat), 160.9344)
		_, _ = hurdle.PredictWithLength(make([]float64, nFeat), -50.0)
		_, _ = hurdle.PredictWithLength(make([]float64, nFeat), math.NaN())
		_, _ = hurdle.PredictWithLength(make([]float64, nFeat), math.Inf(1))
	})
}
