package rtc

import (
	"github.com/pion/logging"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	LogEntry *logrus.Entry
}

func (l Logger) Trace(msg string) {
	l.LogEntry.Trace(msg)
}

func (l Logger) Tracef(format string, args ...interface{}) {
	l.LogEntry.Tracef(format, args...)
}

func (l Logger) Debug(msg string) {
	l.LogEntry.Debug(msg)
}

func (l Logger) Debugf(format string, args ...interface{}) {
	l.LogEntry.Debugf(format, args...)
}

func (l Logger) Info(msg string) {
	l.LogEntry.Info(msg)
}

func (l Logger) Infof(format string, args ...interface{}) {
	l.LogEntry.Infof(format, args...)
}

func (l Logger) Warn(msg string) {
	l.LogEntry.Warn(msg)
}

func (l Logger) Warnf(format string, args ...interface{}) {
	l.LogEntry.Warnf(format, args...)
}

func (l Logger) Error(msg string) {
	l.LogEntry.Error(msg)
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.LogEntry.Errorf(format, args...)
}

type LoggerFactory struct {
	LogEntry *logrus.Entry
}

func (l LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return Logger{LogEntry: l.LogEntry.WithField("scope", scope)}
}
