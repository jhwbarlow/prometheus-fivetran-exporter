package connector

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	pausedEnumGauge = metrics.NewEnumGauge(metrics.BooleanMetricsGaugeValues,
		"Current paused state of a connector")
)
