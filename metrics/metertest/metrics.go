package metertest

import (
	"context"
	"errors"

	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type Meter interface {
	metrics.Meter
	Collect() ([]metrics.Metric, error)
}

type meter struct {
	internalMeter internal.Meter
	reader        metric.Reader
}

func NewMock() (Meter, error) {
	resources := internal.GetResources()

	manualReader := metric.NewManualReader()

	provider := metric.NewMeterProvider(metric.WithResource(resources), metric.WithReader(manualReader))
	otel.SetMeterProvider(provider)

	otelMeter := provider.Meter("ha-influx-grafana")

	return meter{
		internalMeter: internal.Meter{
			OtelMeter: otelMeter,
		},
		reader: manualReader,
	}, nil
}

func (m meter) NewGauge(metric metrics.Metric) error {
	return m.internalMeter.NewGauge(metric)
}

func (m meter) Collect() ([]metrics.Metric, error) {
	var resourceMetrics metricdata.ResourceMetrics

	err := m.reader.Collect(context.Background(), &resourceMetrics)
	if err != nil {
		return nil, err
	}

	var response []metrics.Metric
	for _, scopeMetric := range resourceMetrics.ScopeMetrics {
		for _, metric := range scopeMetric.Metrics {
			switch v := metric.Data.(type) {
			case metricdata.Gauge[float64]:
				for _, datapoint := range v.DataPoints {
					attributes := make(map[string]interface{})
					for _, keyValue := range datapoint.Attributes.ToSlice() {
						attributes[string(keyValue.Key)] = keyValue.Value.AsString()
					}

					response = append(response, metrics.Metric{
						Name:       metric.Name,
						Unit:       metric.Unit,
						Value:      datapoint.Value,
						Attributes: attributes,
					})
				}
			default:
				return nil, errors.New("unknown metric type")
			}
		}
	}

	return response, nil
}
