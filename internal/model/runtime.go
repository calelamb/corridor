package model

// Predictor predicts a scalar target value (such as expected collision rate,
// count, or risk score) from an input vector of numerical features.
type Predictor interface {
	Predict(features []float64) (float64, error)
}
