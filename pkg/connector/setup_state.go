package connector

import (
	"encoding/json"
	"fmt"
)

/*
The current setup state of the connector. The available values are:
incomplete - the setup config is incomplete, the setup tests never succeeded;
connected - the connector is properly set up;
broken - the connector setup config is broken.
*/
type SetupState string

const (
	SetupStateBroken     SetupState = "broken"
	SetupStateConnected  SetupState = "connected"
	SetupStateIncomplete SetupState = "incomplete"
)

func NewSetupState(str string) (SetupState, error) {
	switch str {
	case string(SetupStateBroken):
		return SetupStateBroken, nil
	case string(SetupStateConnected):
		return SetupStateConnected, nil
	case string(SetupStateIncomplete):
		return SetupStateIncomplete, nil
	default:
		return SetupState(""), fmt.Errorf("illegal SetupState: %q", str)
	}
}

func (ss *SetupState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("unmarshalling SetupState: %w", err)
	}

	setupState, err := NewSetupState(str)
	if err != nil {
		return fmt.Errorf("constructing SetupState: %w", err)
	}

	*ss = setupState
	return nil
}
