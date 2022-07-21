package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
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

type EnvVarSourcer struct {
	logger *zap.SugaredLogger
}

func NewEnvVarSourcer(logger *zap.SugaredLogger) *EnvVarSourcer {
	logger = getComponentLogger(logger, "env-var-sourcer")
	return &EnvVarSourcer{logger}
}

func (s *EnvVarSourcer) APIKey() (string, error) {
	return s.getEnvVar(apiKeyEnvVar)
}

func (s *EnvVarSourcer) APISecret() (string, error) {
	return s.getEnvVar(apiSecretEnvVar)
}

func (s *EnvVarSourcer) APICallTimeout() (time.Duration, error) {
	timeoutStr, err := s.getEnvVar(timeoutEnvVar)
	if err != nil {
		return 0, err
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		s.logger.Errorw("parsing API call timeout", "timeout", timeoutStr, "error", err)
		return 0, fmt.Errorf("parsing API call timeout %q: %w", timeoutStr, err)
	}

	return timeout, nil
}

func (s *EnvVarSourcer) CollectedGroupNames() ([]string, error) {
	csv, err := s.getEnvVar(groupsEnvVar)
	if err != nil {
		return nil, err
	}

	split := strings.Split(csv, ",")
	trimmedNames := make([]string, 0, len(split))
	for _, name := range split {
		trimmed := strings.Trim(name, " ")
		if trimmed == "" {
			s.logger.Errorw("invalid group name in environment variable", "name", groupsEnvVar)
			return nil, fmt.Errorf("invalid group name in environment variable %q", groupsEnvVar)
		}

		trimmedNames = append(trimmedNames, trimmed)
	}

	return trimmedNames, nil
}

func (s *EnvVarSourcer) MetricsPort() (uint16, error) {
	portStr, err := s.getEnvVar(metricsPortEnvVar)
	if err != nil {
		return 0, err
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		s.logger.Errorw("parsing metrics port", "port", portStr, "error", err)
		return 0, fmt.Errorf("parsing metrics port %q: %w", portStr, err)
	}

	return uint16(port), nil
}

func (s *EnvVarSourcer) getEnvVar(name string) (string, error) {
	if val := os.Getenv(name); val != "" {
		s.logger.Infow("read environment variable", "name", name)
		return val, nil
	}

	s.logger.Errorw("environment variable not set", "name", name)
	return "", fmt.Errorf("environment variable %q not set", name)
}
