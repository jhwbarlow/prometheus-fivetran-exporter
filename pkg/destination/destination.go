package destination

type Destination struct {
	ID          string
	Name        string
	GroupID     string
	GroupName   string
	Service     string
	SetupStatus SetupStatus
}

type SetupStatus string

const (
	SetupStatusBroken     SetupStatus = "broken"
	SetupStatusConnected  SetupStatus = "connected"
	SetupStatusIncomplete SetupStatus = "incomplete"
)
