package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	apiKeyEnvVar      = "FIVETRAN_API_KEY"
	apiSecretEnvVar   = "FIVETRAN_API_SECRET"
	timeoutEnvVar     = "FIVETRAN_API_CALL_TIMEOUT"
	groupsEnvVar      = "FIVETRAN_COLLECTED_GROUPIDS_CSV"
	metricsPortEnvVar = "METRICS_PORT"
)

type Sourcer interface {
	APIKey() (string, error)
	APISecret() (string, error)
	APICallTimeout() (time.Duration, error)
	CollectedGroupNames() ([]string, error)
	MetricsPort() (uint16, error)
}

type EnvVarSourcer struct{}

func NewEnvVarSourcer() *EnvVarSourcer {
	return new(EnvVarSourcer)
}

func (*EnvVarSourcer) APIKey() (string, error) {
	return getEnvVar(apiKeyEnvVar)
}

func (*EnvVarSourcer) APISecret() (string, error) {
	return getEnvVar(apiSecretEnvVar)
}

func (*EnvVarSourcer) APICallTimeout() (time.Duration, error) {
	timeoutStr, err := getEnvVar(timeoutEnvVar)
	if err != nil {
		return 0, err
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0, fmt.Errorf("parsing API call timeout: %w", err)
	}

	return timeout, nil
}

func (*EnvVarSourcer) CollectedGroupNames() ([]string, error) {
	csv, err := getEnvVar(groupsEnvVar)
	if err != nil {
		return nil, err
	}

	split := strings.Split(csv, ",")
	trimmedNames := make([]string, 0, len(split))
	for _, name := range split {
		trimmed := strings.Trim(name, " ")
		if trimmed == "" {
			return nil, fmt.Errorf("invalid group name in environment variable %q", groupsEnvVar)
		}

		trimmedNames = append(trimmedNames, trimmed)
	}

	return trimmedNames, nil
}

func (*EnvVarSourcer) MetricsPort() (uint16, error) {
	portStr, err := getEnvVar(metricsPortEnvVar)
	if err != nil {
		return 0, err
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parsing metrics port: %w", err)
	}

	return uint16(port), nil
}

func getEnvVar(name string) (string, error) {
	if val := os.Getenv(name); val != "" {
		return val, nil
	}

	return "", fmt.Errorf("environment variable %q not set", name)
}
