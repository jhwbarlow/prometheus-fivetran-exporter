package destination

import (
	"sync"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/metrics"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/destination"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type collectFunc func(*destination.Destination, chan<- prometheus.Metric, *sync.WaitGroup)

const (
	namespace = "fivetran"
	subsystem = "destination"

	gaugeSetupStatusName   = "setup_status"
	gaugeInfoName          = "info"
	counterErrorsTotalName = "errors_total"
)

var (
	gaugeSetupStateFQName = prometheus.BuildFQName(namespace, subsystem, gaugeSetupStatusName)
	gaugeSetupStateDesc   = prometheus.NewDesc(
		gaugeSetupStateFQName,
		setupStatusEnumGauge.Describe(),
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeInfoFQName = prometheus.BuildFQName(namespace, subsystem, gaugeInfoName)
	gaugeInfoDesc   = prometheus.NewDesc(
		gaugeInfoFQName,
		infoEnumGauge.Describe(),
		[]string{"group_name", "group_id", "name", "id", "service"},
		prometheus.Labels{})
)

type Collector struct {
	Describers         []destination.Describer
	counterErrorsTotal *prometheus.CounterVec
	collectFuncs       []collectFunc
	logger             *zap.SugaredLogger
}

func NewCollector(logger *zap.SugaredLogger, describers []destination.Describer) *Collector {
	logger = getComponentLogger(logger, "collector")

	counterErrorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      counterErrorsTotalName,
		Help:      "Total errors encountered querying destination",
	},
		[]string{"group_name"})
	prometheus.MustRegister(counterErrorsTotal)

	for _, describer := range describers {
		// Initialise the error counter to zero for all group names
		counterErrorsTotal.WithLabelValues(describer.GetGroupName()).Add(0)
	}

	collector := &Collector{
		Describers:         describers,
		counterErrorsTotal: counterErrorsTotal,
		logger:             logger,
	}

	collectFuncs := []collectFunc{
		collector.collectSetupStatus,
		collector.collectInfo,
	}
	collector.collectFuncs = collectFuncs

	return collector
}

func (c *Collector) Describe(descsChan chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, descsChan)
}

func (c *Collector) Collect(metricsChan chan<- prometheus.Metric) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(c.Describers))
	for _, describer := range c.Describers {
		go c.collectForDescriber(describer, metricsChan, waitGroup)
	}
	waitGroup.Wait()
}

func (c *Collector) collectForDescriber(describer destination.Describer,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	destination, err := describer.Describe()
	if err != nil {
		// We do not have to send this metric on the metricsChan as it is already registered
		// (it is a metric _belonging_ to this collector, rather than _collected_)
		c.counterErrorsTotal.WithLabelValues(
			describer.GetGroupName()).Inc() // `group_name` label
		c.logger.Errorw("describing destination", "group_name", describer.GetGroupName(), "error", err)
		return
	}

	collectFuncWaitGroup := new(sync.WaitGroup)
	collectFuncWaitGroup.Add(len(c.collectFuncs))
	for _, collectFunc := range c.collectFuncs {
		go collectFunc(destination, metricsChan, collectFuncWaitGroup)
		// TODO: Handle errors and increment error counter.
		// Currently the collectFuncs do not return an error, as
		// we trust that the data they work on is 100% legit, as
		// it was sanity-checked by the describer already
	}
	collectFuncWaitGroup.Wait()
}

func (c *Collector) collectSetupStatus(dest *destination.Destination,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	value := setupStatusGaugeValueConnected
	switch dest.SetupStatus {
	case destination.SetupStatusIncomplete:
		value = setupStatusGaugeValueIncomplete
	case destination.SetupStatusBroken:
		value = setupStatusGaugeValueBroken
	}

	metricsChan <- prometheus.MustNewConstMetric(gaugeSetupStateDesc,
		prometheus.GaugeValue,
		value.GaugeValue(),
		dest.GroupName, // `group_name` label
		dest.Name)      // `name` label

	c.logger.Infow("collected metric",
		"group_name", dest.GroupName,
		"name", dest.Name,
		"metric", gaugeSetupStateFQName)
}

func (c *Collector) collectInfo(dest *destination.Destination,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	metricsChan <- prometheus.MustNewConstMetric(gaugeInfoDesc,
		prometheus.GaugeValue,
		metrics.EnumGaugeValuePresent.GaugeValue(),
		dest.GroupName, // `group_name` label
		dest.GroupID,   // `group_id` label
		dest.Name,      // `name` label
		dest.ID,        // `id` label
		dest.Service)   // `service` label

	c.logger.Infow("collected metric",
		"group_name", dest.GroupName,
		"group_id", dest.GroupID,
		"name", dest.Name,
		"id", dest.ID,
		"service", dest.Service,
		"metric", gaugeInfoFQName)
}
