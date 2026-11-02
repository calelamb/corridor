package model

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"

	"github.com/dmitryikh/leaves"
)

var (
	// versionRegex matches LightGBM format version headers such as version=v4, version=v5.
	// leaves natively accepts version=v2 and version=v3. Because the tree grammar remains
	// unchanged across v3 and v4+, normalizing to v3 allows leaves to parse models exported
	// by modern LightGBM 4.x without error.
	versionRegex = regexp.MustCompile(`\bversion=v[4-9][0-9]*\b`)
)

// LeavesModel wraps a gradient boosted decision tree ensemble loaded via leaves.
type LeavesModel struct {
	ensemble *leaves.Ensemble
}

// LoadLeavesModel loads a LightGBM model from text format via the pure Go leaves parser.
// Compiles and runs cleanly with CGO_ENABLED=0 (zero CGo dependencies).
func LoadLeavesModel(reader io.Reader) (p Predictor, err error) {
	defer func() {
		if r := recover(); r != nil {
			p = nil
			err = fmt.Errorf("panic parsing LightGBM ensemble: %v", r)
		}
	}()

	if reader == nil {
		return nil, fmt.Errorf("nil model reader")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read leaves model: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty leaves model content")
	}

	// Normalize modern LightGBM 4+ version header to v3 for leaves compatibility
	normalized := versionRegex.ReplaceAll(data, []byte("version=v3"))

	bufReader := bufio.NewReader(bytes.NewReader(normalized))
	ensemble, err := leaves.LGEnsembleFromReader(bufReader, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LightGBM ensemble: %w", err)
	}

	return &LeavesModel{
		ensemble: ensemble,
	}, nil
}

// Predict calculates the single-row prediction for the input feature vector.
func (m *LeavesModel) Predict(features []float64) (pred float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			pred = 0
			err = fmt.Errorf("panic in prediction: %v", r)
		}
	}()

	expected := m.ensemble.NFeatures()
	if len(features) < expected {
		return 0, fmt.Errorf("insufficient features: got %d, expected at least %d", len(features), expected)
	}

	nOutputs := m.ensemble.NOutputGroups()
	if nOutputs < 1 {
		nOutputs = 1
	}

	preds := make([]float64, nOutputs)
	if err := m.ensemble.Predict(features, 0, preds); err != nil {
		return 0, fmt.Errorf("prediction failed: %w", err)
	}

	return preds[0], nil
}

// NFeatures returns the number of input features expected by the model.
func (m *LeavesModel) NFeatures() int {
	return m.ensemble.NFeatures()
}

// Name returns the underlying model identifier (e.g. "lightgbm.gbdt").
func (m *LeavesModel) Name() string {
	return m.ensemble.Name()
}
