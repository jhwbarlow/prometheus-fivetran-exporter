package connector

type Connector struct {
	ID                string
	Name              string
	GroupID           string
	GroupName         string
	Service           string
	Paused            bool
	IsHistoricalSync  bool
	SyncFrequencyMins int
	TaskCount         int
	WarningCount      int
	SetupState        SetupState
	SyncState         SyncState
	UpdateState       UpdateState
}

type SetupState string

const (
	SetupStateBroken     SetupState = "broken"
	SetupStateConnected  SetupState = "connected"
	SetupStateIncomplete SetupState = "incomplete"
)

type SyncState string

const (
	SyncStateScheduled   SyncState = "scheduled"
	SyncStateSyncing     SyncState = "syncing"
	SyncStatePaused      SyncState = "paused"
	SyncStateRescheduled SyncState = "rescheduled"
)

type UpdateState string

const (
	UpdateStateOnSchedule UpdateState = "on_schedule"
	UpdateStateDelayed    UpdateState = "delayed"
)
