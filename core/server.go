package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/labstack/echo/middleware"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	pluginEcho "github.com/urunimi/gorest/plugin/echo"
)

// Server provides methods for controlling server's lifecycle
type Server interface {
	Init() error
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
	Logger().Warnf("error: %s", err.Error())
	del.engine.DefaultHTTPErrorHandler(err, c)
}

type server struct {
	services       []App
	sentry         *raven.Client
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

func (s *server) Init() error {
	s.initErrorHandlers()
	s.initLogger()
	s.initReporters()
	for _, svc := range s.services {
		if err := svc.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) initLogger() {
	_logger = getLogger(false)
	if logLevel := os.Getenv("log_level"); logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err == nil {
			_logger.SetLevel(lvl)
		}
	}
	if logFormat := os.Getenv("log_format"); logFormat == "text" {
		_logger.Formatter = &logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 03:04:05",
		}
	} else {
		_logger.Formatter = &logrus.JSONFormatter{}
	}
	s.driver.Logger = pluginEcho.Logger{Logger: _logger}
}

func (s *server) initReporters() error {
	if dsn := os.Getenv("sentry_dsn"); dsn != "" {
		if err := raven.SetDSN(dsn); err != nil {
			return err
		}
		s.sentry = raven.DefaultClient
		registerSentryHook(_logger, s.sentry)
		s.errorListeners = append(s.errorListeners, pluginEcho.NewSentryErrorHandler(s.sentry))
	}
	return nil
}

func (s *server) Start() {
	port := 8080
	if portStr := os.Getenv("port"); portStr != "" {
		port, _ = strconv.Atoi(portStr)
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
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
