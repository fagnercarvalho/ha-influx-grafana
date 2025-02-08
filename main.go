package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/fagnercarvalho/ha-influx-grafana/ha"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics/meter"
)

var UnitOfMeasurementAttribute = "unit_of_measurement"

func main() {
	homeAssistantURL := os.Getenv("HA_URL")
	homeAssistantToken := os.Getenv("HA_TOKEN")
	otelCollectorURL := os.Getenv("OTEL_COLLECTOR_URL")

	ctx := context.Background()

	meter, err := meter.New(otelCollectorURL)
	if err != nil {
		panic(err)
	}

	homeAssistant := ha.NewHomeAssistant(homeAssistantURL, homeAssistantToken)

	filter := ha.State{
		Attributes: map[string]interface{}{
			ha.StateClassAttribute:  ha.StateClassMeasurement,
			ha.DeviceClassAttribute: ha.DeviceClassMoisture,
		},
	}

	states, err := homeAssistant.GetStates(ctx, filter)
	if err != nil {
		panic(err)
	}

	for _, state := range states {
		err := addMetric(ctx, state.EntityID, homeAssistant, meter)
		if err != nil {
			panic(err)
		}
	}

	startHTTPServer()
}

func startHTTPServer() {
	http.HandleFunc("/_status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

/*
query to get air quality greater than x

from(bucket: "influx/autogen")
  |> range(start: -1m)
  |> last()
  |> filter(fn: (r) => r._field != "start_time_unix_nano" and r["unit_of_measurement"] == "µg/m³" and r._value > 300)
*/
