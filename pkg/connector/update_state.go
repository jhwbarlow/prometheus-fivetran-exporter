package connector

import (
	"encoding/json"
	"fmt"
)

/*
The current data update state of the connector. The available values are:
on_schedule - the sync is running smoothly, no delays;
delayed - the data is delayed for a longer time than expected for the update.
*/
type UpdateState string

const (
	UpdateStateOnSchedule UpdateState = "on_schedule"
	UpdateStateDelayed    UpdateState = "delayed"
)

func NewUpdateState(str string) (UpdateState, error) {
	switch str {
	case string(UpdateStateOnSchedule):
		return UpdateStateOnSchedule, nil
	case string(UpdateStateDelayed):
		return UpdateStateDelayed, nil
	default:
		return UpdateState(""), fmt.Errorf("illegal UpdateState: %q", str)
	}
}

func (us *UpdateState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	updateState, err := NewUpdateState(str)
	if err != nil {
		return fmt.Errorf("constructing UpdateState: %w", err)
	}

	*us = updateState
	return nil
}
