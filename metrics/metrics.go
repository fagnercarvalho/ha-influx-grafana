package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Meter struct {
	otelMeter api.Meter
}

type Metric struct {
	Name       string
	Value      float64
	Unit       string
	Attributes map[string]interface{}
	GetValue   func() float64
}

type Gauge struct {
	otelGauge api.Float64ObservableGauge
	otelMeter api.Meter
}

func New(otelCollectorURL string) (Meter, error) {
	ctx := context.Background()

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("ha-influx-grafana"),
		semconv.ServiceVersionKey.String("v0.0.0"),
	)

	otlpExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(otelCollectorURL), otlpmetrichttp.WithInsecure())
	if err != nil {
		return Meter{}, err
	}

	provider := metric.NewMeterProvider(metric.WithResource(resources), metric.WithReader(metric.NewPeriodicReader(otlpExporter, metric.WithInterval(time.Second*30))))
	otel.SetMeterProvider(provider)

	meter := provider.Meter("ha-influx-grafana")

	return Meter{
		otelMeter: meter,
	}, nil
}

func (m Meter) NewGauge(metric Metric) (Gauge, error) {
	var otelAttrs []attribute.KeyValue
	for key, value := range metric.Attributes {
		switch parsedType := value.(type) {
		case string:
			otelAttrs = append(otelAttrs, attribute.String(key, parsedType))
		case float64:
			otelAttrs = append(otelAttrs, attribute.Float64(key, parsedType))
		}
	}

	gauge, err := m.otelMeter.Float64ObservableGauge(metric.Name, api.WithUnit(metric.Unit))
	if err != nil {
		return Gauge{}, err
	}

	_, err = m.otelMeter.RegisterCallback(func(_ context.Context, o api.Observer) error {
		o.ObserveFloat64(gauge, metric.GetValue(), api.WithAttributes(otelAttrs...))

		return nil
	}, gauge)
	if err != nil {
		return Gauge{}, err
	}

	return Gauge{
		otelGauge: gauge,
		otelMeter: m.otelMeter,
	}, nil
}
