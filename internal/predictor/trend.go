package predictor

import "time"

// Direction describes trend movement.
type Direction string

const (
	TrendStable Direction = "STABLE"

	TrendIncreasing Direction = "INCREASING"

	TrendDecreasing Direction = "DECREASING"
)

// Trend represents calculated metric movement.
type Trend struct {
	Metric string

	Direction Direction

	Current float64

	Previous float64

	Delta float64

	Rate float64

	Window time.Duration
}
