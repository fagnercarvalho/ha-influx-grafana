package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/fagnercarvalho/ha-influx-grafana/ha"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
)

var (
	UnitOfMeasurementAttribute = "unit_of_measurement"

	ErrParseState = errors.New("error while trying to parse state")
)

func main() {
	homeAssistantURL := os.Getenv("HA_URL")
	homeAssistantToken := os.Getenv("HA_TOKEN")
	otelCollectorURL := os.Getenv("OTEL_COLLECTOR_URL")

	ctx := context.Background()

	meter, err := metrics.New(otelCollectorURL)
	if err != nil {
		panic(err)
	}

	homeAssistant := ha.NewHomeAssistant(homeAssistantURL, homeAssistantToken)

	filter := ha.State{
		Attributes: map[string]interface{}{
			ha.StateClassAttribute: ha.StateClassMeasurement,
		},
	}

	states, err := homeAssistant.GetStates(ctx, filter)
	if err != nil {
		panic(err)
	}

	for _, state := range states {
		newState := state
		getMetric := func() (metrics.Metric, error) {
			state, err := homeAssistant.GetStateByEntityID(ctx, newState.EntityID)
			if err != nil {
				return metrics.Metric{}, err
			}

			metric, err := convertToMetric(state)
			if err != nil {
				return metrics.Metric{}, err
			}

			return metric, nil
		}

		metric, err := getMetric()
		if err != nil {
			panic(err)
		}

		metric.GetValue = func() float64 {
			metric, err := getMetric()
			if err != nil {
				panic(err)
			}

			return metric.Value
		}

		_, err = meter.NewGauge(metric)
		if err != nil {
			panic(err)
		}
	}

	startHTTPServer()
}

func convertToMetric(state ha.State) (metrics.Metric, error) {
	parsedState, err := strconv.ParseFloat(state.State, 64)
	if err != nil {
		return metrics.Metric{}, errors.Join(ErrParseState, err)
	}

	metric := metrics.Metric{
		Name:  state.EntityID,
		Value: parsedState,
		// https://ucum.nlm.nih.gov/ucum-lhc/
		// https://ucum.org/ucum#para-curly
		Attributes: map[string]interface{}{},
	}

	for attribute, value := range state.Attributes {
		metric.Attributes[attribute] = value

		if attribute == UnitOfMeasurementAttribute {
			value, ok := value.(string)
			if ok {
				metric.Unit = convertUnitToUCUM(value)
			}
		}
	}

	fmt.Println(metric)

	return metric, nil
}

func convertUnitToUCUM(unit string) string {
	switch unit {
	case "°C":
		return "Cel"
	case "µg/m³":
		return "ug/m3"
	case "%":
		return unit
	}

	return unit
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
