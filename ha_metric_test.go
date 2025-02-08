package main

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/fagnercarvalho/ha-influx-grafana/ha"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics/metertest"
)

//go:embed testdata/state.json
var state string

func TestAddMetric(t *testing.T) {
	tests := []struct {
		name          string
		expectMetrics []metrics.Metric
	}{
		{
			name: "happy path",
			expectMetrics: []metrics.Metric{
				{
					Name:  "binary_sensor.motion_sensor",
					Value: 0,
					Unit:  "",
					Attributes: map[string]interface{}{
						"device_class":  "motion",
						"friendly_name": "Motion sensor",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMeter, err := metertest.NewMock()
			if err != nil {
				t.Fatal(err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(state))
			}))

			homeAssistant := ha.NewHomeAssistant(server.URL, "")

			err = addMetric(context.Background(), "entityID", homeAssistant, mockMeter)
			if err != nil {
				t.Fatal(err)
			}

			metrics, err := mockMeter.Collect()
			if err != nil {
				t.Fatal(err)
			}

			if len(metrics) != len(test.expectMetrics) {
				t.Errorf("got %d metrics, want %d", len(metrics), len(test.expectMetrics))
			}

			if !reflect.DeepEqual(test.expectMetrics, metrics) {
				t.Errorf("got metrics %v, want %v", metrics, test.expectMetrics)
			}
		})
	}
}
