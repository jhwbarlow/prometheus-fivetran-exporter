package connector

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	syncStateGaugeValueSyncing     = metrics.NewEnumGaugeValue("syncing", 0)
	syncStateGaugeValueScheduled   = metrics.NewEnumGaugeValue("scheduled", 1)
	syncStateGaugeValueRescheduled = metrics.NewEnumGaugeValue("rescheduled", 2)
	syncStateGaugeValuePaused      = metrics.NewEnumGaugeValue("paused", 3)

	syncStateEnumGauge = metrics.NewEnumGauge([]*metrics.EnumGaugeValue{
		syncStateGaugeValueSyncing,
		syncStateGaugeValueScheduled,
		syncStateGaugeValueRescheduled,
		syncStateGaugeValuePaused,
	},
		"Current sync state of a connector")
)
