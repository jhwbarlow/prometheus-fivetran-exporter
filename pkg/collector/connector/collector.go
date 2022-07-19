package connector

import (
	"fmt"
	"log"
	"sync"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	connectorLister "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/lister/connector"
	groupResolver "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/resolver/group"
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
	Lister connectorLister.Lister

	counterErrorsTotal *prometheus.CounterVec
	collectFuncs       []collectFunc
}

func NewConnectorCollector(lister connectorLister.Lister,
	groupResolver groupResolver.Resolver) (*ConnectorCollector, error) {
	// TODO: By statically resolving the group ID to group Name in the constructor,
	// we will not pick up if the group name changes. To solve this we need to dynamically
	// create the error counter in each scrape (store the count locally in this object)
	// and resolve the value on each scrape, as is done with the connectors
	groupName, err := groupResolver.ResolveIDToName(lister.GroupID())
	if err != nil {
		return nil, fmt.Errorf("resolving group ID %q to group name", lister.GroupID())
	}

	counterErrorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      counterErrorsTotalName,
		Help:      "Total errors encountered querying connectors",
	},
		[]string{"group_name"})

	prometheus.MustRegister(counterErrorsTotal)
	counterErrorsTotal.WithLabelValues(groupName).Add(0)

	collector := &ConnectorCollector{
		Lister:             lister,
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
	// TODO: Change to use Group Name to match definition in constructor where it is initialised to zero.
	// However, if the resolver dynamically looks up the connector name and that errors, we can not then
	// increment for that error as we do not know the name.
	// Thinking about it, we should probably use the group ID for the error metric as it is immutable, but
	// then again, in the config we pass a list of group names, not IDs so that may be a pointless concern
	connectors, err := c.Lister.List()
	if err != nil {
		// TODO: I do not believe we have to send this metric on the metricsChan as it
		// is already registered
		c.counterErrorsTotal.WithLabelValues(
			c.Lister.GroupID()).Inc() // `group_id` label
		log.Printf("Error getting connectors: %v", err) // TODO: Use logger
		return
	}

	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(c.collectFuncs))
	for _, collectFunc := range c.collectFuncs {
		go collectFunc(connectors, metricsChan, waitGroup)
	}
	waitGroup.Wait()

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
