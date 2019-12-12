package core

import (
	"os"

	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	pluginLogrus "github.com/urunimi/gorest/plugin/logrus"
)

//Gets application _logger
func getLogger(newLogger bool) *logrus.Logger {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 03:04:05",
	}
	var log *logrus.Logger
	if newLogger {
		log = logrus.New()
	} else {
		log = logrus.StandardLogger()
	}
	log.Formatter = formatter
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
	return log
}

func registerSentryHook(log *logrus.Logger, client *raven.Client) {
	hook := pluginLogrus.NewSentryHook(client, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	log.Hooks.Add(hook)
}

// Customize _logger from config
func initLogger(log *logrus.Logger, config map[string]string) {
	if config["level"] != "" {
		lvl, err := logrus.ParseLevel(config["level"])
		if err == nil {
			log.SetLevel(lvl)
		}
	}

}
