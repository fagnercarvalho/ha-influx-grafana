package ha

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	StateClassAttribute   = "state_class"
	StateClassMeasurement = "measurement"

	DeviceClassAttribute = "device_class"
	DeviceClassMoisture  = "moisture"

	ErrHomeAssistantRequest       = errors.New("error when trying to get states from Home Assistant API")
	ErrHomeAssistantReadBody      = errors.New("error when trying to read body from Home Assistant API response")
	ErrHomeAssistantJSONUnmarshal = errors.New("error when trying to unmarshal JSON from Home Assistant API response")
)

type HomeAssistant struct {
	serverURL string
	token     string
}

type State struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastUpdated time.Time              `json:"last_updated"`
}

func NewHomeAssistant(serverURL, token string) HomeAssistant {
	return HomeAssistant{serverURL: serverURL, token: token}
}

func (ha HomeAssistant) GetStates(ctx context.Context, f State) ([]State, error) {
	client := http.Client{}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, ha.serverURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Bearer "+ha.token)

	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Join(ErrHomeAssistantRequest, err)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Join(ErrHomeAssistantReadBody, err)
	}

	var states []State
	err = json.Unmarshal(bytes, &states)
	if err != nil {
		return nil, errors.Join(ErrHomeAssistantJSONUnmarshal, err)
	}

	states = filter(states, f)

	return states, nil
}

func (ha HomeAssistant) GetStateByEntityID(ctx context.Context, entityID string) (State, error) {
	client := http.Client{}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, ha.serverURL+"/"+entityID, nil)
	if err != nil {
		return State{}, err
	}

	request.Header.Add("Authorization", "Bearer "+ha.token)

	response, err := client.Do(request)
	if err != nil {
		return State{}, errors.Join(ErrHomeAssistantRequest, err)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return State{}, errors.Join(ErrHomeAssistantReadBody, err)
	}

	var state State
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return State{}, errors.Join(ErrHomeAssistantJSONUnmarshal, err)
	}

	return state, nil
}

func filter(states []State, filter State) []State {
	if filter.Attributes == nil || len(filter.Attributes) == 0 {
		return states
	}

	var filteredStates []State

	for _, state := range states {
		var matchingAttribute bool
		for key, value := range filter.Attributes {
			if state.Attributes[key] == value {
				matchingAttribute = true
			}
		}

		if matchingAttribute {
			filteredStates = append(filteredStates, state)
		}
	}

	return filteredStates
}
