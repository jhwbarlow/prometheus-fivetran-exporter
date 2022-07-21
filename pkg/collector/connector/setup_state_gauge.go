package connector

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	setupStateGaugeValueConnected  = metrics.NewEnumGaugeValue("connected", 0)
	setupStateGaugeValueIncomplete = metrics.NewEnumGaugeValue("incomplete", 1)
	setupStateGaugeValueBroken     = metrics.NewEnumGaugeValue("broken", 2)

	setupStateEnumGauge = metrics.NewEnumGauge([]*metrics.EnumGaugeValue{
		setupStateGaugeValueConnected,
		setupStateGaugeValueIncomplete,
		setupStateGaugeValueBroken,
	}, "Current setup state of a connector")
)
