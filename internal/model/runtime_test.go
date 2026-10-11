package model_test

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"corridor/internal/model"
)

type referenceData struct {
	FeatureNames                  []string    `json:"feature_names"`
	TestRows                      [][]float64 `json:"test_rows"`
	SegmentLengths                []float64   `json:"segment_lengths"`
	LightGBMPredictions           []float64   `json:"lightgbm_predictions"`
	HurdlePredictionsDefaultLength []float64   `json:"hurdle_predictions_default_length"`
	HurdlePredictionsActualLength  []float64   `json:"hurdle_predictions_actual_length"`
}

func loadReferenceData(t *testing.T) referenceData {
	t.Helper()
	data, err := os.ReadFile("testdata/reference_predictions.json")
	if err != nil {
		t.Fatalf("failed to read reference predictions: %v", err)
	}
	var ref referenceData
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatalf("failed to unmarshal reference predictions: %v", err)
	}
	return ref
}

func TestLoadLeavesModel_MatchesPythonReference(t *testing.T) {
	ref := loadReferenceData(t)

	f, err := os.Open("testdata/lightgbm_poisson.txt")
	if err != nil {
		t.Fatalf("failed to open lightgbm test model: %v", err)
	}
	defer f.Close()

	predictor, err := model.LoadLeavesModel(f)
	if err != nil {
		t.Fatalf("LoadLeavesModel failed: %v", err)
	}

	// Verify Predictor interface
	var _ model.Predictor = predictor

	const tolerance = 1e-5
	for i, row := range ref.TestRows {
		pred, err := predictor.Predict(row)
		if err != nil {
			t.Fatalf("row %d Predict failed: %v", i, err)
		}
		expected := ref.LightGBMPredictions[i]
		diff := math.Abs(pred - expected)
		if diff > tolerance {
			t.Errorf("row %d: got prediction %v, expected %v, diff %v > tolerance %v",
				i, pred, expected, diff, tolerance)
		}
	}
}

func TestLoadLeavesModel_InsufficientFeatures(t *testing.T) {
	f, err := os.Open("testdata/lightgbm_poisson.txt")
	if err != nil {
		t.Fatalf("failed to open lightgbm test model: %v", err)
	}
	defer f.Close()

	predictor, err := model.LoadLeavesModel(f)
	if err != nil {
		t.Fatalf("LoadLeavesModel failed: %v", err)
	}

	// Model expects 10 features, provide only 2
	shortFeatures := []float64{1.0, 2.0}
	_, err = predictor.Predict(shortFeatures)
	if err == nil {
		t.Fatal("expected error on insufficient features, got nil")
	}
}

func TestLoadLeavesModel_InvalidInput(t *testing.T) {
	badReaders := []struct {
		name   string
		reader *strings.Reader
	}{
		{"empty reader", strings.NewReader("")},
		{"corrupted header", strings.NewReader("not a lightgbm model")},
	}

	for _, tc := range badReaders {
		t.Run(tc.name, func(t *testing.T) {
			_, err := model.LoadLeavesModel(tc.reader)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestLoadHurdleModel_MatchesPythonReference(t *testing.T) {
	ref := loadReferenceData(t)

	f, err := os.Open("testdata/hurdle_model.json")
	if err != nil {
		t.Fatalf("failed to open hurdle test model: %v", err)
	}
	defer f.Close()

	predictor, err := model.LoadHurdleModel(f)
	if err != nil {
		t.Fatalf("LoadHurdleModel failed: %v", err)
	}

	// Verify Predictor interface
	var _ model.Predictor = predictor

	const tolerance = 1e-5

	// Test 1: Predict with default length (features slice matching p feature names)
	t.Run("DefaultLength", func(t *testing.T) {
		for i, row := range ref.TestRows {
			pred, err := predictor.Predict(row)
			if err != nil {
				t.Fatalf("row %d Predict failed: %v", i, err)
			}
			expected := ref.HurdlePredictionsDefaultLength[i]
			diff := math.Abs(pred - expected)
			if diff > tolerance {
				t.Errorf("row %d: got prediction %v, expected %v, diff %v > tolerance %v",
					i, pred, expected, diff, tolerance)
			}
		}
	})

	// Test 2: Predict with actual segment length passed as last feature (p+1 elements)
	t.Run("ActualLength", func(t *testing.T) {
		for i, row := range ref.TestRows {
			rowWithLength := append([]float64(nil), row...)
			rowWithLength = append(rowWithLength, ref.SegmentLengths[i])

			pred, err := predictor.Predict(rowWithLength)
			if err != nil {
				t.Fatalf("row %d Predict failed: %v", i, err)
			}
			expected := ref.HurdlePredictionsActualLength[i]
			diff := math.Abs(pred - expected)
			if diff > tolerance {
				t.Errorf("row %d: got prediction %v, expected %v, diff %v > tolerance %v",
					i, pred, expected, diff, tolerance)
			}
		}
	})
}

func TestLoadHurdleModel_InsufficientFeatures(t *testing.T) {
	f, err := os.Open("testdata/hurdle_model.json")
	if err != nil {
		t.Fatalf("failed to open hurdle test model: %v", err)
	}
	defer f.Close()

	predictor, err := model.LoadHurdleModel(f)
	if err != nil {
		t.Fatalf("LoadHurdleModel failed: %v", err)
	}

	shortFeatures := []float64{1.0, 2.0}
	_, err = predictor.Predict(shortFeatures)
	if err == nil {
		t.Fatal("expected error on insufficient features, got nil")
	}
}

func TestLoadHurdleModel_InvalidInput(t *testing.T) {
	badInputs := []struct {
		name string
		data []byte
	}{
		{"empty input", []byte("")},
		{"malformed json", []byte("{not json}")},
		{"invalid model type", []byte(`{"model_type": "random_forest"}`)},
	}

	for _, tc := range badInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := model.LoadHurdleModel(bytes.NewReader(tc.data))
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}
