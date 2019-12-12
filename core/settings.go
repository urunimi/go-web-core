package core

import (
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
)

type settings struct {
	config *viper.Viper
	sentry *raven.Client
}

func (s *settings) init() (err error) {
	if err = s.initConfig(); err != nil {
		return
	}
	if err = s.initSentry(); err != nil {
		return
	}
	return
}

func (s *settings) initConfig() (err error) {
	config := viper.New()
	defaultEnv := "local"
	config.SetDefault("server_env", defaultEnv)
	config.AutomaticEnv()
	s.config = config
	return
}

func (s *settings) initSentry() (err error) {
	if !s.config.IsSet("sentry_dsn") {
		return
	}
	err = raven.SetDSN(s.config.GetString("sentry_dsn"))
	s.sentry = raven.DefaultClient
	return
}
