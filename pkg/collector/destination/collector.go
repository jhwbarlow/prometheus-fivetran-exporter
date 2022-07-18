package destination

import (
	"log"
	"sync"

	destinationDescriber "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/describer/destination"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/destination"
	"github.com/prometheus/client_golang/prometheus"
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
	gaugeSetupStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeSetupStatusName),
		"Current setup state of a destination",
		[]string{"group_name", "name"},
		prometheus.Labels{})
	gaugeInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, gaugeInfoName),
		"Information about a destination",
		[]string{"group_name", "group_id", "name", "id", "service"},
		prometheus.Labels{})
)

type DestinationCollector struct {
	Describer          destinationDescriber.Describer
	counterErrorsTotal *prometheus.CounterVec
	collectFuncs       []collectFunc
}

func NewDestinationCollector(describer destinationDescriber.Describer) *DestinationCollector {
	// TODO: group_name not ID!
	counterErrorsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      counterErrorsTotalName,
		Help:      "Total errors encountered querying destination",
	},
		[]string{"group_name"})

	prometheus.MustRegister(counterErrorsTotal)
	counterErrorsTotal.WithLabelValues(describer.GroupID()).Add(0)

	collector := &DestinationCollector{
		Describer:          describer,
		counterErrorsTotal: counterErrorsTotal,
	}

	collectFuncs := []collectFunc{
		collector.collectSetupStatus,
		collector.collectInfo,
	}
	collector.collectFuncs = collectFuncs

	return collector
}

func (c *DestinationCollector) Describe(descsChan chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, descsChan) // TODO: Review usage in light of extra label values may pop into and out of existence
}

func (c *DestinationCollector) Collect(metricsChan chan<- prometheus.Metric) {
	destination, err := c.Describer.Describe()
	if err != nil {
		c.counterErrorsTotal.WithLabelValues(
			c.Describer.GroupID()) // `group_id` label
		log.Printf("Error getting destination: %v", err) // TODO: Use logger
		return
	}

	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(c.collectFuncs))
	for _, collectFunc := range c.collectFuncs {
		go collectFunc(destination, metricsChan, waitGroup)
	}
	waitGroup.Wait()

	// TODO: Handle errors and increment error counter
}

func (c *DestinationCollector) collectSetupStatus(dest *destination.Destination,
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
		float64(value),
		dest.GroupID, // `group` label
		dest.Name)    // `destination_name` label

}

func (c *DestinationCollector) collectInfo(dest *destination.Destination,
	metricsChan chan<- prometheus.Metric,
	waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	metricsChan <- prometheus.MustNewConstMetric(gaugeInfoDesc,
		prometheus.GaugeValue,
		float64(infoGaugeValuePresent),
		dest.GroupName, // `group_name` label
		dest.GroupID,   // `group_id` label
		dest.Name,      // `name` label
		dest.ID,        // `id` label
		dest.Service)   // `service` label

}
