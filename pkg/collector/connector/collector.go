package connector

import (
	"log"
	"sync"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	connectorLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/connector"
	"github.com/prometheus/client_golang/prometheus"
)

type collectFunc func([]*connector.Connector, chan<- prometheus.Metric, *sync.WaitGroup)

const (
	namespace = "fivetran"
	subsystem = "connector"

	gaugePausedName           = "paused"
	gaugeSetupStateName       = "setup_state"
	gaugeSyncStateName        = "sync_state"
	gaugeUpdateStateName      = "update_state"
	gaugeInfoName             = "info"
	gaugeSyncFrequencyName    = "sync_frequency_seconds" // NOTE: Prometheus recommended time unit, and not Hz!
	gaugeTaskCountName        = "task_count"
	gaugeWarningCountName     = "warning_count"
	gaugeInHistoricalSyncName = "in_historical_sync"
	counterErrorsTotalName    = "errors_total"
)

var (
	// TODO: Add enum gauge values to description
	gaugePausedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugePausedName),
		"Current paused state of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeSetupStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeSetupStateName),
		"Current setup state of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeSyncStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeSyncStateName),
		"Current sync state of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeUpdateStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeUpdateStateName),
		"Current update state of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeSyncFrequencyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeSyncFrequencyName),
		"Current sync frequency of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	// TODO: Would this be better served by a counter? We'd have to keep locally a map
	// of the last warning count for each connector, and then increment the count by
	// any difference between one scrape and the next
	gaugeTaskCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeTaskCountName),
		"Number of outstanding tasks for a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	// TODO: Would this be better served by a counter? We'd have to keep locally a map
	// of the last warning count for each connector, and then increment the count by
	// any difference between one scrape and the next
	gaugeWarningCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeWarningCountName),
		"Number of current warnings/alerts for a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeInHistoricalSync = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeInHistoricalSyncName),
		"Current historical sync state of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeInfoName),
		"Information about a connector",
		[]string{"group_name", "group_id", "name", "id", "service"},
		prometheus.Labels{})
)

type ConnectorCollector struct {
	Listers []connectorLister.Lister

	counterErrorsTotal *prometheus.CounterVec
	collectFuncs       []collectFunc
}

func NewConnectorCollector(listers []connectorLister.Lister) (*ConnectorCollector, error) {
	counterErrorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      counterErrorsTotalName,
		Help:      "Total errors encountered querying connectors",
	},
		[]string{"group_name"})
	prometheus.MustRegister(counterErrorsTotal)

	for _, lister := range listers {
		// Initialise the error counter to zero for all group names
		counterErrorsTotal.WithLabelValues(lister.GetGroupName()).Add(0)
	}

	collector := &ConnectorCollector{
		Listers:            listers,
		counterErrorsTotal: counterErrorsTotal,
	}

	collectFuncs := []collectFunc{
		collector.collectPaused,
		collector.collectSetupState,
		collector.collectSyncState,
		collector.collectUpdateState,
		collector.collectInfo,
		collector.collectSyncFrequency,
		collector.collectTaskCount,
		collector.collectWarningCount,
		collector.collectInHistoricalSync,
	}
	collector.collectFuncs = collectFuncs

	return collector, nil
}

func (c *ConnectorCollector) Describe(descsChan chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, descsChan)
}

func (c *ConnectorCollector) Collect(metricsChan chan<- prometheus.Metric) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(c.Listers))
	for _, lister := range c.Listers {
		go c.collectForLister(lister, metricsChan, waitGroup)
	}
	waitGroup.Wait()
}

func (c *ConnectorCollector) collectForLister(lister connectorLister.Lister,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	connectors, err := lister.List()
	if err != nil {
		// We do not have to send this metric on the metricsChan as it is already registered
		// (it is a metric _belonging_ to this collector, rather than _collected_)
		c.counterErrorsTotal.WithLabelValues(
			lister.GetGroupName()).Inc() // `group_name` label
		log.Printf("Error listing connectors: %v", err) // TODO: Use logger
		return
	}

	collectFuncWaitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(c.collectFuncs))
	for _, collectFunc := range c.collectFuncs {
		go collectFunc(connectors, metricsChan, waitGroup)
	}
	collectFuncWaitGroup.Wait()

	// TODO: Handle errors and increment error counter
}

func (c *ConnectorCollector) collectPaused(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := pausedGaugeValueFalse
		if conn.Paused {
			value = pausedGaugeValueTrue
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugePausedDesc,
			prometheus.GaugeValue,
			float64(value),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}

func (c *ConnectorCollector) collectSetupState(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := setupStateGaugeValueConnected
		switch conn.SetupState {
		case connector.SetupStateIncomplete:
			value = setupStateGaugeValueIncomplete
		case connector.SetupStateBroken:
			value = setupStateGaugeValueBroken
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugeSetupStateDesc,
			prometheus.GaugeValue,
			float64(value),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}

func (c *ConnectorCollector) collectSyncState(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := syncStateGaugeValueSyncing
		switch conn.SyncState {
		case connector.SyncStateScheduled:
			value = syncStateGaugeValueScheduled
		case connector.SyncStateRescheduled:
			value = syncStateGaugeValueRescheduled
		case connector.SyncStatePaused:
			value = syncStateGaugeValuePaused
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugeSyncStateDesc,
			prometheus.GaugeValue,
			float64(value),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}

func (c *ConnectorCollector) collectUpdateState(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := updateStateGaugeValueOnSchedule
		if conn.UpdateState == connector.UpdateStateDelayed {
			value = updateStateGaugeValueDelayed
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugeUpdateStateDesc,
			prometheus.GaugeValue,
			float64(value),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}

func (c *ConnectorCollector) collectInfo(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		metricsChan <- prometheus.MustNewConstMetric(gaugeInfoDesc,
			prometheus.GaugeValue,
			float64(infoGaugeValuePresent),
			conn.GroupName, // `group_name` label
			conn.GroupID,   // `group_id` label
			conn.Name,      // `name` label
			conn.ID,        // `id` label
			conn.Service)   // `service` label
	}
}

func (c *ConnectorCollector) collectSyncFrequency(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		metricsChan <- prometheus.MustNewConstMetric(gaugeSyncFrequencyDesc,
			prometheus.GaugeValue,
			float64(conn.SyncFrequencyMins*60), // Convert to seconds
			conn.GroupName,                     // `group_name` label
			conn.Name)                          // `name` label
	}
}

func (c *ConnectorCollector) collectWarningCount(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		metricsChan <- prometheus.MustNewConstMetric(gaugeWarningCount,
			prometheus.GaugeValue,
			float64(conn.WarningCount),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}

func (c *ConnectorCollector) collectTaskCount(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		metricsChan <- prometheus.MustNewConstMetric(gaugeTaskCount,
			prometheus.GaugeValue,
			float64(conn.TaskCount),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}

func (c *ConnectorCollector) collectInHistoricalSync(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := inHistoricalSyncGaugeValueFalse
		if conn.Paused {
			value = inHistoricalSyncGaugeValueTrue
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugeInHistoricalSync,
			prometheus.GaugeValue,
			float64(value),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label
	}
}
