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
