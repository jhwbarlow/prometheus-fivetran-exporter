package destination

type Destination struct {
	ID          string
	Name        string
	GroupID     string
	GroupName   string
	Service     string
	SetupStatus SetupStatus
}
