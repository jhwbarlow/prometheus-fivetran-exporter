package connector

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	updateStateGaugeValueOnSchedule = metrics.NewEnumGaugeValue("on_schedule", 0)
	updateStateGaugeValueDelayed    = metrics.NewEnumGaugeValue("delayed", 1)

	updateStateEnumGauge = metrics.NewEnumGauge([]*metrics.EnumGaugeValue{
		updateStateGaugeValueOnSchedule,
		updateStateGaugeValueDelayed,
	},
		"Current update state of a connector")
)
