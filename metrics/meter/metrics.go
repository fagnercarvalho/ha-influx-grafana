package meter

import (
	"context"
	"time"

	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

func New(otelCollectorURL string) (metrics.Meter, error) {
	ctx := context.Background()

	resources := internal.GetResources()

	otlpExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(otelCollectorURL), otlpmetrichttp.WithInsecure())
	if err != nil {
		return internal.Meter{}, err
	}

	provider := metric.NewMeterProvider(metric.WithResource(resources), metric.WithReader(metric.NewPeriodicReader(otlpExporter, metric.WithInterval(time.Second*30))))
	otel.SetMeterProvider(provider)

	otelMeter := provider.Meter("ha-influx-grafana")

	return internal.Meter{
		OtelMeter: otelMeter,
	}, nil
}
