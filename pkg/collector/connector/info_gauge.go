package connector

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	infoEnumGauge = metrics.NewEnumGauge(metrics.PresentMetricsGaugeValues,
		"Information about a connector")
)
