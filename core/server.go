package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	pluginEcho "github.com/urunimi/go-web-core/plugin/echo"
)

// Server provides methods for controlling application lifecycle
type Server interface {
	Init(configPath string, config interface{}) error
	Start()
	Exit(sig os.Signal)
}

type App interface {
	Init() error
	RegisterRoute(driver *Engine)
	Clean() error
}

var (
	_logger *logrus.Logger
	// loggers contains applications loggers
	_loggers = map[string]*logrus.Logger{}
)

type Context = echo.Context

type Engine = echo.Echo

type server struct {
	services   []App
	settings   *settings
	driver     *Engine
	httpServer *http.Server
}

type reqValidator struct {
	validator *validator.Validate
}

func Logger() *logrus.Logger {
	if _logger == nil {
		_logger = logrus.StandardLogger()
	}
	return _logger
}

func NewEngine() *Engine {
	return echo.New()
}

func (cv *reqValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (s *server) Init(configPath string, configStruct interface{}) error {
	s.driver.Validator = &reqValidator{validator: validator.New()}
	s.initSettings(configPath, configStruct)
	s.initLoggers()
	s.initReporters()
	for _, svc := range s.services {
		if err := svc.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) initReporters() {
	if s.settings.sentry != nil {
		registerSentryHook(_logger, s.settings.sentry)
		pluginEcho.SetErrorHandlerForSentry(s.driver, s.settings.sentry)
	}
}

func (s *server) initSettings(configPath string, configStruct interface{}) {
	s.settings = &settings{}
	if err := s.settings.init(configPath, configStruct); err != nil {
		panic(err)
	}
}

func (s *server) initLoggers() {
	config := s.settings.config
	if config.IsSet("_logger") {
		_logger = getLogger(false)
		config := config.GetStringMapString("_logger")
		initLogger(_logger, config) // default _logger
	}
	if config.IsSet("_loggers") {
		// Multiple _loggers
		loggers := config.GetStringMap("_loggers")
		for k := range loggers {
			config := config.GetStringMapString("_loggers." + k)
			_loggers[k] = getLogger(true)
			initLogger(_loggers[k], config)
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

	s.driver.Use(middleware.Recover())
	s.driver.Use(middleware.CORS())

	for _, svc := range s.services {
		svc.RegisterRoute(s.driver)
	}
	s.registerExitHandler()
	_ = s.httpServer.ListenAndServe()
}

func (s *server) Exit(sig os.Signal) {
	for _, svc := range s.services {
		_ = svc.Clean()
	}
	// Shutdown http server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		_logger.Fatal("Server Shutdown Error:", err)
	}
	// Exit
	if sig != nil {
		os.Exit(2) // SIGINT
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
		driver:   NewEngine(),
	}
	return Server(server)
}
