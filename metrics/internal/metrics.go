package internal

import (
	"context"

	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Meter struct {
	OtelMeter api.Meter
}

func (m Meter) NewGauge(metric metrics.Metric) error {
	var otelAttrs []attribute.KeyValue
	for key, value := range metric.Attributes {
		switch parsedType := value.(type) {
		case string:
			otelAttrs = append(otelAttrs, attribute.String(key, parsedType))
		case float64:
			otelAttrs = append(otelAttrs, attribute.Float64(key, parsedType))
		}
	}

	gauge, err := m.OtelMeter.Float64ObservableGauge(metric.Name, api.WithUnit(metric.Unit))
	if err != nil {
		return err
	}

	_, err = m.OtelMeter.RegisterCallback(func(_ context.Context, o api.Observer) error {
		o.ObserveFloat64(gauge, metric.GetValue(), api.WithAttributes(otelAttrs...))

		return nil
	}, gauge)
	if err != nil {
		return err
	}

	return nil
}

func GetResources() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("ha-influx-grafana"),
		semconv.ServiceVersionKey.String("v0.0.0"),
	)
}
