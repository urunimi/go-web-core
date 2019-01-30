package core

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
)

// Server provides methods for controlling application lifecycle
type Server interface {
	Init(configPath string, config interface{})
	Start()
	Exit(sig os.Signal)
}

type App interface {
	Init()
	RegisterRoute(driver *Engine)
	Clean() error
}

var (
	Logger *logrus.Logger
	// Loggers contains applications loggers
	Loggers = map[string]*logrus.Logger{}
)

type Context = echo.Context

type Engine = echo.Echo

type server struct {
	services   []App
	settings   *settings
	driver     *Engine
	httpServer *http.Server
}

func (s *server) Init(configPath string, configStruct interface{}) {
	s.initSettings(configPath, configStruct)
	s.initLoggers()
	for _, svc := range s.services {
		svc.Init()
	}
	return
}

func (s *server) initSettings(configPath string, configStruct interface{}) {
	s.settings = &settings{}
	if err := s.settings.init(configPath, configStruct); err != nil {
		panic(err)
	}
}

func (s *server) initLoggers() {
	config := s.settings.config
	if config.IsSet("logger") {
		Logger = getLogger(false)
		config := config.GetStringMapString("logger")
		initLogger(Logger, config) // default logger
		if s.settings.sentry != nil {
			registerSentryHook(Logger, s.settings.sentry)
		}
	}
	if config.IsSet("loggers") {
		// Multiple loggers
		loggers := config.GetStringMap("loggers")
		for k := range loggers {
			config := config.GetStringMapString("loggers." + k)
			Loggers[k] = getLogger(true)
			initLogger(Loggers[k], config)
		}
	}
}

func (s *server) Start() {
	config := s.settings.config
	config.SetDefault("server.port", 8080)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GetInt("server.port")),
		Handler: s.driver,
	}
	for _, svc := range s.services {
		svc.RegisterRoute(s.driver)
	}
	s.registerExitHandler()
	if err := s.httpServer.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (s *server) Exit(sig os.Signal) {
	for _, svc := range s.services {
		_ = svc.Clean()
	}
}

func (s *server) registerExitHandler() {
	// Wait for interrupt signal to gracefully shutdown the server
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		s.Exit(sig)
	}()
}

func NewServer(services ...App) Server {
	server := &server{
		services: services,
		driver:   echo.New(),
	}
	return Server(server)
}
