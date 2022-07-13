package destination

import (
	"encoding/json"
	"fmt"
)

type SetupStatus string

const (
	SetupStatusBroken     SetupStatus = "broken"
	SetupStatusConnected  SetupStatus = "connected"
	SetupStatusIncomplete SetupStatus = "incomplete"
)

func NewSetupStatus(str string) (SetupStatus, error) {
	switch str {
	case string(SetupStatusBroken):
		return SetupStatusBroken, nil
	case string(SetupStatusConnected):
		return SetupStatusConnected, nil
	case string(SetupStatusIncomplete):
		return SetupStatusIncomplete, nil
	default:
		return SetupStatus(""), fmt.Errorf("illegal SetupStatus: %q", str)
	}
}

func (ss *SetupStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	setupStatus, err := NewSetupStatus(str)
	if err != nil {
		return fmt.Errorf("constructing SetupStatus: %w", err)
	}

	*ss = setupStatus
	return nil
}
