package model

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// hurdleJSON defines the serialized JSON format exported by corridor_ml.models.export.
type hurdleJSON struct {
	ModelType    string   `json:"model_type"`
	Version      string   `json:"version"`
	FeatureNames []string `json:"feature_names"`
	Scaler       struct {
		Means map[string]float64 `json:"means"`
		Stds  map[string]float64 `json:"stds"`
	} `json:"scaler"`
	HurdleStage struct {
		Distribution string             `json:"distribution"`
		Intercept    float64            `json:"intercept"`
		Coefficients map[string]float64 `json:"coefficients"`
	} `json:"hurdle_stage"`
	CountStage struct {
		Distribution string             `json:"distribution"`
		Intercept    float64            `json:"intercept"`
		Dispersion   float64            `json:"dispersion"`
		Coefficients map[string]float64 `json:"coefficients"`
	} `json:"count_stage"`
	Exposure struct {
		Column           string  `json:"column"`
		DefaultLengthM   float64 `json:"default_length_m"`
		ReferenceLengthM float64 `json:"reference_length_m"`
		LogOffset        bool    `json:"log_offset"`
	} `json:"exposure"`
}

// HurdleModel implements two-stage Hurdle Negative Binomial GLM inference in pure Go.
// Stage 1: Logistic regression modeling P(Y > 0) to account for observer reporting bias.
// Stage 2: Zero-truncated Negative Binomial modeling E(Y | Y > 0) with exposure offset.
// Overall Expected Rate: E(Y) = P(Y > 0) * E(Y | Y > 0).
type HurdleModel struct {
	featureNames         []string
	means                []float64
	stds                 []float64
	hurdleIntercept      float64
	hurdleCoefs          []float64
	countIntercept       float64
	countCoefs           []float64
	dispersion           float64
	defaultLengthM       float64
	referenceLengthM     float64
	hasExposureLogOffset bool
}

// LoadHurdleModel deserializes a Hurdle Negative Binomial model from JSON and returns
// a Predictor implementation executing pure Go vector arithmetic with zero CGo.
func LoadHurdleModel(reader io.Reader) (Predictor, error) {
	if reader == nil {
		return nil, fmt.Errorf("nil model reader")
	}

	var data hurdleJSON
	dec := json.NewDecoder(reader)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode hurdle model json: %w", err)
	}

	if data.ModelType != "hurdle_negative_binomial" {
		return nil, fmt.Errorf("invalid model_type %q: expected %q", data.ModelType, "hurdle_negative_binomial")
	}

	p := len(data.FeatureNames)
	if p == 0 {
		return nil, fmt.Errorf("no feature names defined in hurdle model")
	}

	means := make([]float64, p)
	stds := make([]float64, p)
	hurdleCoefs := make([]float64, p)
	countCoefs := make([]float64, p)

	for i, feat := range data.FeatureNames {
		mean, ok := data.Scaler.Means[feat]
		if !ok {
			return nil, fmt.Errorf("missing scaler mean for feature %q", feat)
		}
		means[i] = mean

		std, ok := data.Scaler.Stds[feat]
		if !ok {
			return nil, fmt.Errorf("missing scaler std for feature %q", feat)
		}
		if std < 1e-9 || math.IsNaN(std) {
			std = 1.0
		}
		stds[i] = std

		hCoef, ok := data.HurdleStage.Coefficients[feat]
		if !ok {
			return nil, fmt.Errorf("missing hurdle coefficient for feature %q", feat)
		}
		hurdleCoefs[i] = hCoef

		cCoef, ok := data.CountStage.Coefficients[feat]
		if !ok {
			return nil, fmt.Errorf("missing count coefficient for feature %q", feat)
		}
		countCoefs[i] = cCoef
	}

	defaultLen := data.Exposure.DefaultLengthM
	if defaultLen <= 0 {
		defaultLen = 160.9344 // ~0.1 mile
	}

	refLen := data.Exposure.ReferenceLengthM
	if refLen <= 0 {
		refLen = 160.9344
	}

	disp := data.CountStage.Dispersion
	if disp <= 0 || math.IsNaN(disp) {
		disp = 1.0
	}

	return &HurdleModel{
		featureNames:         append([]string(nil), data.FeatureNames...),
		means:                means,
		stds:                 stds,
		hurdleIntercept:      data.HurdleStage.Intercept,
		hurdleCoefs:          hurdleCoefs,
		countIntercept:       data.CountStage.Intercept,
		countCoefs:           countCoefs,
		dispersion:           disp,
		defaultLengthM:       defaultLen,
		referenceLengthM:     refLen,
		hasExposureLogOffset: data.Exposure.LogOffset,
	}, nil
}

// Predict calculates expected collision rate E(Y) = P(Y > 0) * E(Y | Y > 0).
// If len(features) == len(feature_names), uses default segment length exposure.
// If len(features) >= len(feature_names) + 1, uses features[len(feature_names)] as segment length in meters.
func (m *HurdleModel) Predict(features []float64) (float64, error) {
	p := len(m.featureNames)
	if len(features) < p {
		return 0, fmt.Errorf("insufficient features: got %d, expected at least %d", len(features), p)
	}

	segLen := m.defaultLengthM
	if len(features) >= p+1 {
		segLen = features[p]
	}

	return m.PredictWithLength(features[:p], segLen)
}

// PredictWithLength calculates expected collision rate with an explicit segment length in meters.
func (m *HurdleModel) PredictWithLength(features []float64, segmentLengthM float64) (float64, error) {
	p := len(m.featureNames)
	if len(features) < p {
		return 0, fmt.Errorf("insufficient features: got %d, expected at least %d", len(features), p)
	}

	// 1. Standardize features & compute linear predictors
	logit := m.hurdleIntercept
	countEta := m.countIntercept

	for i := 0; i < p; i++ {
		std := m.stds[i]
		if std < 1e-9 || math.IsNaN(std) {
			std = 1.0
		}
		xScaled := (features[i] - m.means[i]) / std
		logit += m.hurdleCoefs[i] * xScaled
		countEta += m.countCoefs[i] * xScaled
	}

	// 2. Stage 1: Logistic Hurdle probability P(Y > 0)
	var pHurdle float64
	if logit >= 0 {
		pHurdle = 1.0 / (1.0 + math.Exp(-logit))
	} else {
		ex := math.Exp(logit)
		pHurdle = ex / (1.0 + ex)
	}

	// 3. Stage 2: Exposure offset and Zero-Truncated Negative Binomial expectation
	if m.hasExposureLogOffset {
		length := math.Max(segmentLengthM, 1e-3)
		refLen := math.Max(m.referenceLengthM, 1e-6)
		offset := math.Log(length / refLen)
		countEta += offset
	}

	// Numerical clamp [-30.0, 30.0] matching training runtime
	if countEta > 30.0 {
		countEta = 30.0
	} else if countEta < -30.0 {
		countEta = -30.0
	}
	mu := math.Exp(countEta)

	disp := math.Max(m.dispersion, 1e-6)
	invAlpha := 1.0 / disp
	logP0 := -invAlpha * math.Log1p(disp*mu)
	denom := -math.Expm1(logP0)
	if denom < 1e-12 {
		denom = 1e-12
	} else if denom > 1.0 {
		denom = 1.0
	}

	condMean := 1.0
	if mu >= 1e-12 {
		condMean = mu / denom
	}
	if condMean < 1.0 {
		condMean = 1.0
	}

	// 4. Overall expected rate
	rate := pHurdle * condMean
	if rate < 0.0 {
		rate = 0.0
	}

	return rate, nil
}

// FeatureNames returns the ordered slice of feature names required by this model.
func (m *HurdleModel) FeatureNames() []string {
	return append([]string(nil), m.featureNames...)
}

// NFeatures returns the number of input features expected by this model.
func (m *HurdleModel) NFeatures() int {
	return len(m.featureNames)
}
