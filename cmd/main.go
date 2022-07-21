package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/blendle/zapdriver"
	connectorcollector "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/connector"
	destinationcollector "github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/collector/destination"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/config"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/connector"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/destination"
	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/group"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	apiURL = "https://api.fivetran.com"
)

func main() {
	// Logging setup
	zapLogger, err := zapdriver.NewProduction()
	if err != nil {
		log.Fatalf("Error constructing zap logger: %v", err)
	}
	logger := zapLogger.Sugar()
	defer logger.Sync() // Flush logs at the end of the application's lifetime

	// Get config from environment
	apiKey, apiSecret, apiTimeout, collectedGroupNames, metricsPort, err := getConfig(logger)
	if err != nil {
		logger.Fatalw("Error sourcing config", "error", err)
	}

	// List the groups so that we can use the resolver to get the ID from the provided names
	groupLister, err := group.NewAPILister(logger, apiKey, apiSecret, apiURL, 10*time.Second)
	if err != nil {
		logger.Fatalw("Error constructing group lister", "error", err)
	}
	// XXX: By statically resolving the group ID to group name in main(),
	// we will not pick up if the group name changes. However, we treat the group name as
	// immutable even if it is technically not, as we specify the group name, not the group
	// ID in the configuration. Using the group ID in the configuration is not as nice, as
	// it requires knowing the non-deterministic ID in advance.
	groupResolver := group.NewGroupListerResolver(logger, groupLister)

	connectorListers := make([]connector.Lister, 0, len(collectedGroupNames))
	destinationDescribers := make([]destination.Describer, 0, len(collectedGroupNames))
	for _, groupName := range collectedGroupNames {
		groupID, err := groupResolver.ResolveNameToID(groupName)
		if err != nil {
			logger.Fatalw("Error resolving group name to ID", "group_name", groupName, "error", err)
		}

		// Construct a connector lister for each listed group
		connectorLister, err := connector.NewAPILister(logger,
			apiKey,
			apiSecret,
			apiURL,
			groupID,
			groupName,
			apiTimeout)
		if err != nil {
			logger.Fatalw("Error constructing connector lister", "group_name", groupName, "error", err)
		}
		connectorListers = append(connectorListers, connectorLister)

		// Construct a destination describer for each listed group
		destinationDescriber, err := destination.NewAPIDescriber(logger,
			apiKey,
			apiSecret,
			apiURL,
			groupID,
			groupName,
			apiTimeout)
		if err != nil {
			logger.Fatalw("Error construction destination describer", "group_name", groupName, "error", err)
		}
		destinationDescribers = append(destinationDescribers, destinationDescriber)
	}

	connectorCollector, err := connectorcollector.NewCollector(logger, connectorListers)
	if err != nil {
		logger.Fatalw("Error constructing connector collector", "error", err)
	}

	destinationCollector := destinationcollector.NewCollector(logger, destinationDescribers)

	prometheus.MustRegister(destinationCollector)
	prometheus.MustRegister(connectorCollector)

	if err := run(logger, metricsPort); err != nil {
		logger.Fatalw("Error running exporter", "error", err)
	}
}

func getConfig(logger *zap.SugaredLogger) (apiKey string,
	apiSecret string,
	apiCallTimeout time.Duration,
	collectedGroupNames []string,
	metricsPort uint16,
	err error) {
	var configSourcer config.Sourcer = config.NewEnvVarSourcer(logger)

	apiKey, err = configSourcer.APIKey()
	if err != nil {
		logger.Errorw("getting API Key from config", "error", err)
		err = fmt.Errorf("getting API Key from config: %w", err)
		return
	}

	apiSecret, err = configSourcer.APISecret()
	if err != nil {
		logger.Errorw("getting API Secret from config", "error", err)
		err = fmt.Errorf("getting API Secret from config: %w", err)
		return
	}

	apiCallTimeout, err = configSourcer.APICallTimeout()
	if err != nil {
		logger.Errorw("getting API call timeout from config", "error", err)
		err = fmt.Errorf("getting API call timeout from config: %w", err)
		return
	}

	collectedGroupNames, err = configSourcer.CollectedGroupNames()
	if err != nil {
		logger.Errorw("getting collected group names from config", "error", err)
		err = fmt.Errorf("getting collected group names from config: %w", err)
		return
	}

	metricsPort, err = configSourcer.MetricsPort()
	if err != nil {
		logger.Errorw("getting metrics port from config", "error", err)
		err = fmt.Errorf("getting metrics port from config: %w", err)
		return
	}

	logger.Infow("got config",
		"api_key", apiKey,
		"api_secret", "<redacted>",
		"api_call_timeout", apiCallTimeout,
		"collected_group_names", collectedGroupNames,
		"metrics_port", metricsPort)
	return
}

func run(logger *zap.SugaredLogger, metricsPort uint16) error {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
		logger.Errorw("running webserver", "port", metricsPort, "error", err)
		return fmt.Errorf("running webserver on port %d: %w", metricsPort, err)
	}

	// Will never get here
	return nil
}
