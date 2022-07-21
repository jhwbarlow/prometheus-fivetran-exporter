package logging

import (
	"sync"

	"go.uber.org/zap"
)

func InitPackageLogger(baseLogger *zap.SugaredLogger,
	packageLogName string,
	lock *sync.Mutex,
	packageLogger **zap.SugaredLogger) {
	lock.Lock()
	defer lock.Unlock()

	if *packageLogger == nil {
		*packageLogger = baseLogger.Named(packageLogName)
	}
}
