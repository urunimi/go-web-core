package echo

import (
	"io"

	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	header, prefix string
}

func (l Logger) SetHeader(h string) {
	l.header = h
}

func (l Logger) Level() log.Lvl {
	switch l.Logger.Level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	case logrus.InfoLevel:
		return log.INFO
	default:
		l.Panic("Invalid level")
	}

	return log.OFF
}

func (l Logger) SetPrefix(s string) {
	l.prefix = s
}

func (l Logger) Prefix() string {
	return l.prefix
}

func (l Logger) SetLevel(lvl log.Lvl) {
	switch lvl {
	case log.DEBUG:
		logrus.SetLevel(logrus.DebugLevel)
	case log.WARN:
		logrus.SetLevel(logrus.WarnLevel)
	case log.ERROR:
		logrus.SetLevel(logrus.ErrorLevel)
	case log.INFO:
		logrus.SetLevel(logrus.InfoLevel)
	default:
		l.Panic("Invalid level")
	}
}

func (l Logger) Output() io.Writer {
	return l.Out
}

func (l Logger) Printj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Print()
}

func (l Logger) Debugj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Debug()
}

func (l Logger) Infoj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Info()
}

func (l Logger) Warnj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Warn()
}

func (l Logger) Errorj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Error()
}

func (l Logger) Fatalj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Fatal()
}

func (l Logger) Panicj(j log.JSON) {
	l.WithFields(logrus.Fields(j)).Panic()
}
