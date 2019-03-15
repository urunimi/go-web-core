package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/middleware"

	"github.com/Sirupsen/logrus"
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	pluginEcho "github.com/urunimi/go-web-core/plugin/echo"
)

// Server provides methods for controlling server's lifecycle
type Server interface {
	Init(configPath string, config interface{}) error
	Start()
	Exit(sig os.Signal)
}

// App provides methods for controlling an app's lifecycle
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

//Context is echo context
type Context = echo.Context

// Engine define http engine
type Engine = echo.Echo

// HTTPError define http error
type HTTPError = echo.HTTPError

type ErrorListener interface {
	OnError(err error, c Context)
}

type defaultErrorListener struct {
	engine *Engine
}

func (del *defaultErrorListener) OnError(err error, c Context) {
	del.engine.DefaultHTTPErrorHandler(err, c)
}

type server struct {
	services       []App
	settings       *settings
	driver         *Engine
	httpServer     *http.Server
	errorListeners []ErrorListener
}

type reqValidator struct {
	validator *validator.Validate
}

// Logger return logger
func Logger() *logrus.Logger {
	if _logger == nil {
		_logger = logrus.StandardLogger()
	}
	return _logger
}

// NewEngine give new http engine
func NewEngine() *Engine {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Validator = &reqValidator{validator: validator.New()}
	return e
}

func (cv *reqValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (s *server) Init(configPath string, configStruct interface{}) error {
	s.initErrorHandlers()
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

func (s *server) initSettings(configPath string, configStruct interface{}) {
	s.settings = &settings{}
	if err := s.settings.init(configPath, configStruct); err != nil {
		panic(err)
	}
}

func (s *server) initLoggers() {
	config := s.settings.config
	if config.IsSet("logger") {
		_logger = getLogger(false)
		config := config.GetStringMapString("logger")
		initLogger(_logger, config)
		s.driver.Logger = pluginEcho.Logger{Logger: _logger}
	}
	if config.IsSet("loggers") {
		// Multiple _loggers
		loggers := config.GetStringMap("loggers")
		for k := range loggers {
			config := config.GetStringMapString("loggers." + k)
			_loggers[k] = getLogger(true)
			initLogger(_loggers[k], config)
		}
	}
}

func (s *server) initReporters() {
	if s.settings.sentry != nil {
		registerSentryHook(_logger, s.settings.sentry)
		s.errorListeners = append(s.errorListeners, pluginEcho.NewSentryErrorHandler(s.driver, s.settings.sentry, _logger))
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

func (s *server) initErrorHandlers() {
	s.errorListeners = []ErrorListener{
		&defaultErrorListener{s.driver},
	}
	s.driver.HTTPErrorHandler = func(e error, i echo.Context) {
		for _, el := range s.errorListeners {
			el.OnError(e, i)
		}
	}
}

// NewServer return new server instance
func NewServer(services ...App) Server {
	server := &server{
		services: services,
		driver:   NewEngine(),
	}
	return Server(server)
}
