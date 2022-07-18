package connector

type pausedGaugeValue int

const (
	pausedGaugeValueFalse pausedGaugeValue = iota
	pausedGaugeValueTrue
)

type inHistoricalSyncGaugeValue int

const (
	inHistoricalSyncGaugeValueFalse pausedGaugeValue = iota
	inHistoricalSyncGaugeValueTrue
)

type setupStateGaugeValue int

const (
	setupStateGaugeValueConnected setupStateGaugeValue = iota
	setupStateGaugeValueIncomplete
	setupStateGaugeValueBroken
)

type syncStateGaugeValue int

const (
	syncStateGaugeValueSyncing syncStateGaugeValue = iota
	syncStateGaugeValueScheduled
	syncStateGaugeValueRescheduled
	syncStateGaugeValuePaused
)

type updateStateGaugeValue int

const (
	updateStateGaugeValueOnSchedule updateStateGaugeValue = iota
	updateStateGaugeValueDelayed
)

type infoGaugeValue int

const (
	infoGaugeValuePresent infoGaugeValue = 1
)
