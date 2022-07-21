package destination

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	setupStatusGaugeValueConnected  = metrics.NewEnumGaugeValue("connected", 0)
	setupStatusGaugeValueIncomplete = metrics.NewEnumGaugeValue("incomplete", 1)
	setupStatusGaugeValueBroken     = metrics.NewEnumGaugeValue("broken", 2)

	setupStatusEnumGauge = metrics.NewEnumGauge([]*metrics.EnumGaugeValue{
		setupStatusGaugeValueConnected,
		setupStatusGaugeValueIncomplete,
		setupStatusGaugeValueBroken,
	}, "Current setup status of a destination")
)
