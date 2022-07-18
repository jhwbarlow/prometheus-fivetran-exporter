package destination

type infoGaugeValue int

const (
	infoGaugeValuePresent infoGaugeValue = 1
)

type setupStatusGaugeValue int

const (
	setupStatusGaugeValueConnected setupStatusGaugeValue = iota
	setupStatusGaugeValueIncomplete
	setupStatusGaugeValueBroken
)
