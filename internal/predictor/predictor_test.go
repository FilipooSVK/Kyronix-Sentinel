package predictor

import (
	"testing"
)

func TestPredictorEvaluate(t *testing.T) {

	p := New()

	result := p.Evaluate(
		nil,
	)

	if result.Risk != RiskLow {

		t.Fatalf(
			"expected low risk",
		)
	}
}
