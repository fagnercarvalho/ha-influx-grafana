package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/fagnercarvalho/ha-influx-grafana/ha"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
	"strconv"
)

var (
	StateClassMeasurement       = "measurement"
	ErrParseState         error = errors.New("error while trying to parse state")
)

func main() {
	// [x] create grafana influx datasource
	// [x] fix influx flow to OTEL
	// [x] filter by state class instead of measurement name
	// [x] push measurements to influx

	// push to Ubuntu server

	// hide secrets in .env and GitHub secrets
	// make sure I can push to GitHub

	// create readme
	// consider making repo public?

	// https://docs.influxdata.com/influxdb/v1.3/concepts/key_concepts/
	// grafana SELECT "gauge" FROM "autogen"."bar"

	serverURL := ""
	token := ""

	ctx := context.Background()

	meter, err := metrics.New()
	if err != nil {
		panic(err)
	}

	homeAssistant := ha.NewHomeAssistant(serverURL, token)
	states, err := homeAssistant.GetStates(ctx)
	if err != nil {
		panic(err)
	}

	filteredStates := filterByStateClass(states, StateClassMeasurement)

	for _, filteredState := range filteredStates {
		newFilteredState := filteredState
		getMetric := func() (metrics.Metric, error) {
			state, err := homeAssistant.GetStateByEntityID(ctx, newFilteredState.EntityID)
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

	fmt.Scanln()
}

func filterByStateClass(states []ha.State, stateClass string) []ha.State {
	var filteredStates []ha.State

	for _, state := range states {
		if state.Attributes["state_class"] == stateClass {
			filteredStates = append(filteredStates, state)
		}
	}

	return filteredStates
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

		if attribute == "unit_of_measurement" {
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
