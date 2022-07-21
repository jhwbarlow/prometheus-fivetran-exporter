package jsonhttp

import (
	"sync"

	"github.com/jhwbarlow/prometheus-fivetran-exporter/pkg/logging"
	"go.uber.org/zap"
)

var (
	_lock          = new(sync.Mutex)
	_packageLogger *zap.SugaredLogger
)

func getPackageLogger(baseLogger *zap.SugaredLogger) *zap.SugaredLogger {
	logging.InitPackageLogger(baseLogger, "jsonhttp", _lock, &_packageLogger)
	return _packageLogger
}

func getComponentLogger(baseLogger *zap.SugaredLogger, componentName string) *zap.SugaredLogger {
	return getPackageLogger(baseLogger).Named(componentName)
}
