package predictor

import "testing"

func TestCalculateRiskMemoryGrowth(t *testing.T) {

	result := CalculateRisk(
		[]Trend{

			{
				Metric: "memory",

				Direction: TrendIncreasing,

				Rate: 8,
			},
		},
	)

	if result.Level != RiskMedium {

		t.Fatalf(
			"expected medium risk, got %s",
			result.Level,
		)
	}

	if result.Score != 30 {

		t.Fatalf(
			"expected score 30, got %d",
			result.Score,
		)
	}

	if len(result.Reasons) == 0 {

		t.Fatal(
			"expected risk reason",
		)
	}
}

func TestCalculateRiskHealthDegradation(
	t *testing.T,
) {

	trends := []Trend{
		{
			Metric:    "health_score",
			Direction: TrendDecreasing,
			Previous:  80,
			Current:   45,
			Delta:     -35,
		},
	}

	assessment := CalculateRisk(
		trends,
	)

	if assessment.Score != 20 {
		t.Fatalf(
			"expected score 20, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskLow {
		t.Fatalf(
			"expected LOW risk, got %s",
			assessment.Level,
		)
	}
}

func TestCalculateRiskHealthConfirmsMemoryRisk(
	t *testing.T,
) {

	trends := []Trend{
		{
			Metric:    "memory",
			Direction: TrendIncreasing,
			Current:   75,
			Rate:      6,
		},
		{
			Metric:    "health_score",
			Direction: TrendDecreasing,
			Previous:  80,
			Current:   45,
		},
	}

	assessment := CalculateRisk(
		trends,
	)

	if assessment.Score != 50 {
		t.Fatalf(
			"expected score 50, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskMedium {
		t.Fatalf(
			"expected MEDIUM risk, got %s",
			assessment.Level,
		)
	}
}
