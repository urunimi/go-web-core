package core

import (
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
)

type settings struct {
	config *viper.Viper
	sentry *raven.Client
}

func (s *settings) init(configPath string, configStruct interface{}) (err error) {
	if err = s.initConfig(configPath, configStruct); err != nil {
		return
	}
	if err = s.initSentry(); err != nil {
		return
	}
	return
}

func (s *settings) initConfig(configPath string, configStruct interface{}) (err error) {
	config := viper.New()
	defaultEnv := "local"
	config.SetDefault("SERVER_ENV", defaultEnv)
	config.AutomaticEnv()
	if configPath != "" {
		config.AddConfigPath(configPath)
	} else {
		config.AddConfigPath("./config/")
	}
	serverEnv := config.GetString("SERVER_ENV")
	config.SetConfigName("config." + serverEnv)
	err = config.ReadInConfig()
	if err != nil {
		return
	}
	config.Set("env", serverEnv)
	s.config = config
	err = config.Unmarshal(configStruct)
	return
}

func (s *settings) initSentry() (err error) {
	if !s.config.IsSet("Sentry") {
		return
	}
	err = raven.SetDSN(s.config.GetString("Sentry.Dsn"))
	s.sentry = raven.DefaultClient
	return
}
