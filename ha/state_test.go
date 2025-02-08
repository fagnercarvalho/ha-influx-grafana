package ha

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
)

//go:embed testdata/states.json
var states string

func TestGetStates(t *testing.T) {
	tests := []struct {
		name           string
		filter         State
		expectCount    int
		expectEntities []string
	}{
		{
			name: "Filter by moisture devices should only return states for entities of this type",
			filter: State{
				Attributes: map[string]interface{}{
					DeviceClassAttribute: DeviceClassMoisture,
				},
			},
			expectCount: 2,
			expectEntities: []string{
				"binary_sensor.water_leakage_sensor",
				"sensor.soil_moisture",
			},
		},
		{
			name: "Filter by measurements states should only return states for entities of this type",
			filter: State{
				Attributes: map[string]interface{}{
					StateClassAttribute: StateClassMeasurement,
				},
			},
			expectCount: 5,
			expectEntities: []string{
				"sensor.temperature_sensor",
				"sensor.humidity_sensor",
				"sensor.air_quality_sensor",
				"sensor.motion_sensor_battery",
				"sensor.soil_moisture",
			},
		},
		{
			name: "Filter by measurements states and moisture devices should only return states for entities of these types",
			filter: State{
				Attributes: map[string]interface{}{
					StateClassAttribute:  StateClassMeasurement,
					DeviceClassAttribute: DeviceClassMoisture,
				},
			},
			expectCount: 6,
			expectEntities: []string{
				"sensor.temperature_sensor",
				"sensor.humidity_sensor",
				"sensor.air_quality_sensor",
				"sensor.motion_sensor_battery",
				"sensor.soil_moisture",
				"binary_sensor.water_leakage_sensor",
			},
		},
		{
			name:        "No filter should return states for all entities",
			filter:      State{},
			expectCount: 8,
			expectEntities: []string{
				"sensor.temperature_sensor",
				"sensor.humidity_sensor",
				"sensor.air_quality_sensor",
				"sensor.motion_sensor_battery",
				"sensor.soil_moisture",
				"binary_sensor.water_leakage_sensor",
				"automation.turn_on_lights",
				"binary_sensor.motion_sensor",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(states))
			}))

			homeAssistant := NewHomeAssistant(server.URL, "")

			states, err := homeAssistant.GetStates(context.Background(), test.filter)
			if err != nil {
				t.Fatal(err)
			}

			if len(states) != test.expectCount {
				t.Errorf("Expected %d states, got %d", test.expectCount, len(states))
			}

			var actualEntities []string
			for _, state := range states {
				actualEntities = append(actualEntities, state.EntityID)
			}

			sort.Strings(actualEntities)
			sort.Strings(test.expectEntities)

			if !reflect.DeepEqual(actualEntities, test.expectEntities) {
				t.Errorf("Expected states %v, got %v", test.expectEntities, actualEntities)
			}
		})
	}
}
