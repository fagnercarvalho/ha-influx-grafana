package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fagnercarvalho/ha-influx-grafana/ha"
	"github.com/fagnercarvalho/ha-influx-grafana/metrics"
)

func addMetric(ctx context.Context, entityID string, homeAssistant ha.HomeAssistant, meter metrics.Meter) error {
	getMetric := func() (metrics.Metric, error) {
		currentState, err := homeAssistant.GetStateByEntityID(ctx, entityID)
		if err != nil {
			return metrics.Metric{}, err
		}

		metric, err := convertToMetric(currentState)
		if err != nil {
			return metrics.Metric{}, err
		}

		return metric, nil
	}

	metric, err := getMetric()
	if err != nil {
		return err
	}

	metric.GetValue = func() float64 {
		metric, err := getMetric()
		if err != nil {
			panic(err)
		}

		return metric.Value
	}

	return meter.NewGauge(metric)
}

func convertToMetric(state ha.State) (metrics.Metric, error) {
	stateAsInt := convertOnOffToInteger(state.State)

	parsedState, err := strconv.ParseFloat(stateAsInt, 64)
	if err != nil {
		fmt.Printf("Error to parse state for %v: %v. Using -1 as state value \n", state.EntityID, err)

		parsedState = -1
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

	fmt.Println("Converted Home Assistant state to OTel metric", metric)

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

func convertOnOffToInteger(state string) string {
	if state == "on" {
		return "1"
	} else if state == "off" {
		return "0"
	}

	return state
}
