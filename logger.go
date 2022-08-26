package stdlib

import (
	"os"
	"sync"

	"go.uber.org/zap"
)

var lock = &sync.Mutex{}

type logger struct {
	logger  *zap.SugaredLogger
	enabled bool
}

var loggerInstance *logger

func getLoggerInstance() *logger {
	if loggerInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if loggerInstance == nil {
			l := ConsoleLogger()
			enabled := true
			if lme := os.Getenv("LOGGER_ENABLED"); lme == "false" {
				enabled = false
			}
			loggerInstance = &logger{
				logger:  l,
				enabled: enabled,
			}
		}
	}
	return loggerInstance
}

func (l *logger) Debug(args ...interface{}) {
	if l.enabled {
		l.logger.Debugf("%s", args...)
	}
}

func (l *logger) Info(args ...interface{}) {
	if l.enabled {
		l.logger.Infof("%s", args...)
	}
}

func (l *logger) Infof(template string, args ...interface{}) {
	if l.enabled {
		l.logger.Infof(template, args...)
	}
}

func (l *logger) Warn(args ...interface{}) {
	if l.enabled {
		l.logger.Warn(args...)
	}
}

func (l *logger) Warnf(template string, args ...interface{}) {
	if l.enabled {
		l.logger.Warnf(template, args...)
	}
}

func (l *logger) Error(args ...interface{}) {
	if l.enabled {
		l.logger.Error(args...)
	}
}

func (l *logger) Errorf(template string, args ...interface{}) {
	if l.enabled {
		l.logger.Errorf(template, args...)
	}
}
