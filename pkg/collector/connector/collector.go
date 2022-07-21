package connector

import (
	"sync"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
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
	gaugePausedFQName = prometheus.BuildFQName(namespace, subsystem, gaugePausedName)
	gaugePausedDesc   = prometheus.NewDesc(
		gaugePausedFQName,
		pausedEnumGauge.Describe(),
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeSetupStateFQName = prometheus.BuildFQName(namespace, subsystem, gaugeSetupStateName)
	gaugeSetupStateDesc   = prometheus.NewDesc(
		gaugeSetupStateFQName,
		setupStateEnumGauge.Describe(),
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeSyncStateFQName = prometheus.BuildFQName(namespace, subsystem, gaugeSyncStateName)
	gaugeSyncStateDesc   = prometheus.NewDesc(
		gaugeSyncStateFQName,
		syncStateEnumGauge.Describe(),
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeUpdateStateFQName = prometheus.BuildFQName(namespace, subsystem, gaugeUpdateStateName)
	gaugeUpdateStateDesc   = prometheus.NewDesc(
		gaugeUpdateStateFQName,
		updateStateEnumGauge.Describe(),
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeSyncFrequencyFQName = prometheus.BuildFQName(namespace, subsystem, gaugeSyncFrequencyName)
	gaugeSyncFrequencyDesc   = prometheus.NewDesc(
		gaugeSyncFrequencyFQName,
		"Current sync frequency of a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	// TODO: Would this be better served by a counter? We'd have to keep locally a map
	// of the last warning count for each connector, and then increment the count by
	// any difference between one scrape and the next
	gaugeTaskCountFQName = prometheus.BuildFQName(namespace, subsystem, gaugeTaskCountName)
	gaugeTaskCount       = prometheus.NewDesc(
		gaugeTaskCountFQName,
		"Number of outstanding tasks for a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	// TODO: Would this be better served by a counter? We'd have to keep locally a map
	// of the last warning count for each connector, and then increment the count by
	// any difference between one scrape and the next
	gaugeWarningCountFQName = prometheus.BuildFQName(namespace, subsystem, gaugeWarningCountName)
	gaugeWarningCount       = prometheus.NewDesc(
		gaugeWarningCountFQName,
		"Number of current warnings/alerts for a connector",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeInHistoricalSyncFQName = prometheus.BuildFQName(namespace, subsystem, gaugeInHistoricalSyncName)
	gaugeInHistoricalSync       = prometheus.NewDesc(
		gaugeInHistoricalSyncFQName,
		inHistoricalSyncEnumGauge.Describe(),
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeInfoFQName = prometheus.BuildFQName(namespace, subsystem, gaugeInfoName)
	gaugeInfo       = prometheus.NewDesc(
		gaugeInfoFQName,
		infoEnumGauge.Describe(),
		[]string{"group_name", "group_id", "name", "id", "service"},
		prometheus.Labels{})
)

type Collector struct {
	Listers []connector.Lister

	counterErrorsTotal *prometheus.CounterVec
	collectFuncs       []collectFunc
	logger             *zap.SugaredLogger
}

func NewCollector(logger *zap.SugaredLogger, listers []connector.Lister) (*Collector, error) {
	logger = getComponentLogger(logger, "collector")

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

	collector := &Collector{
		Listers:            listers,
		counterErrorsTotal: counterErrorsTotal,
		logger:             logger,
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

func (c *Collector) Describe(descsChan chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, descsChan)
}

func (c *Collector) Collect(metricsChan chan<- prometheus.Metric) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(c.Listers))
	for _, lister := range c.Listers {
		go c.collectForLister(lister, metricsChan, waitGroup)
	}
	waitGroup.Wait()
}

func (c *Collector) collectForLister(lister connector.Lister,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	connectors, err := lister.List()
	if err != nil {
		// We do not have to send this metric on the metricsChan as it is already registered
		// (it is a metric _belonging_ to this collector, rather than _collected_)
		c.counterErrorsTotal.WithLabelValues(
			lister.GetGroupName()).Inc() // `group_name` label
		c.logger.Errorw("listing connectors", "group_name", lister.GetGroupName(), "error", err)
		return
	}

	collectFuncWaitGroup := new(sync.WaitGroup)
	collectFuncWaitGroup.Add(len(c.collectFuncs))
	for _, collectFunc := range c.collectFuncs {
		go collectFunc(connectors, metricsChan, collectFuncWaitGroup)
		// TODO: Handle errors and increment error counter.
		// Currently the collectFuncs do not return an error, as
		// we trust that the data they work on is 100% legit, as
		// it was sanity-checked by the lister already
	}
	collectFuncWaitGroup.Wait()
}

func (c *Collector) collectPaused(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := metrics.EnumGaugeValueFalse
		if conn.Paused {
			value = metrics.EnumGaugeValueTrue
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugePausedDesc,
			prometheus.GaugeValue,
			value.GaugeValue(),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugePausedFQName)
	}
}

func (c *Collector) collectSetupState(connectors []*connector.Connector,
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
			value.GaugeValue(),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeSetupStateFQName)
	}
}

func (c *Collector) collectSyncState(connectors []*connector.Connector,
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
			value.GaugeValue(),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeSyncStateFQName)
	}
}

func (c *Collector) collectUpdateState(connectors []*connector.Connector,
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
			value.GaugeValue(),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeUpdateStateFQName)
	}
}

func (c *Collector) collectInfo(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		metricsChan <- prometheus.MustNewConstMetric(gaugeInfo,
			prometheus.GaugeValue,
			metrics.EnumGaugeValuePresent.GaugeValue(),
			conn.GroupName, // `group_name` label
			conn.GroupID,   // `group_id` label
			conn.Name,      // `name` label
			conn.ID,        // `id` label
			conn.Service)   // `service` label

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"group_id", conn.GroupID,
			"name", conn.Name,
			"id", conn.ID,
			"service", conn.Service,
			"metric", gaugeInfoFQName)
	}
}

func (c *Collector) collectSyncFrequency(connectors []*connector.Connector,
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

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeSyncFrequencyFQName)
	}
}

func (c *Collector) collectWarningCount(connectors []*connector.Connector,
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

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeWarningCountFQName)
	}
}

func (c *Collector) collectTaskCount(connectors []*connector.Connector,
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

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeTaskCountFQName)
	}
}

func (c *Collector) collectInHistoricalSync(connectors []*connector.Connector,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Create one gauge metric per connector
	for _, conn := range connectors {
		value := metrics.EnumGaugeValueFalse
		if conn.Paused {
			value = metrics.EnumGaugeValueTrue
		}

		metricsChan <- prometheus.MustNewConstMetric(gaugeInHistoricalSync,
			prometheus.GaugeValue,
			value.GaugeValue(),
			conn.GroupName, // `group_name` label
			conn.Name)      // `name` label

		c.logger.Infow("collected metric",
			"group_name", conn.GroupName,
			"name", conn.Name,
			"metric", gaugeInHistoricalSyncFQName)
	}
}
