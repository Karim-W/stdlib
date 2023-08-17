package httpclient

import "log"

type Logger interface {
	Debug(args ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Debug(args ...interface{}) {
	log.Println(args...)
}
