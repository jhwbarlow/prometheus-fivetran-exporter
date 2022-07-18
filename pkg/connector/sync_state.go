package connector

import (
	"encoding/json"
	"fmt"
)

/*
The current sync state of the connector. The available values are:
scheduled - the sync is waiting to be run;
syncing - the sync is currently running;
paused - the sync is currently paused;
rescheduled - the sync is waiting until more API calls are available in the source service.
*/
type SyncState string

const (
	SyncStateScheduled   SyncState = "scheduled"
	SyncStateSyncing     SyncState = "syncing"
	SyncStatePaused      SyncState = "paused"
	SyncStateRescheduled SyncState = "rescheduled"
)

func NewSyncState(str string) (SyncState, error) {
	switch str {
	case string(SyncStateScheduled):
		return SyncStateScheduled, nil
	case string(SyncStateSyncing):
		return SyncStateSyncing, nil
	case string(SyncStatePaused):
		return SyncStatePaused, nil
	case string(SyncStateRescheduled):
		return SyncStateRescheduled, nil
	default:
		return SyncState(""), fmt.Errorf("illegal SyncState: %q", str)
	}
}

func (ss *SyncState) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("unmarshalling SyncState: %w", err)
	}

	syncState, err := NewSyncState(str)
	if err != nil {
		return fmt.Errorf("constructing SyncState: %w", err)
	}

	*ss = syncState
	return nil
}
