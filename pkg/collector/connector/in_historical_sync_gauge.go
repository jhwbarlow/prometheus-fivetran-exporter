package connector

import "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"

var (
	inHistoricalSyncEnumGauge = metrics.NewEnumGauge(metrics.BooleanMetricsGaugeValues,
		"Whether or not a connector is currently performing a historical sync")
)
