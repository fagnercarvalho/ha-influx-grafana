package metrics

type Metric struct {
	Name       string
	Value      float64
	Unit       string
	Attributes map[string]interface{}
	GetValue   func() float64
}

type Meter interface {
	NewGauge(metric Metric) error
}
