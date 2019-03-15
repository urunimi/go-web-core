package echo

import (
	"github.com/Sirupsen/logrus"
	"github.com/getsentry/raven-go"
	"github.com/labstack/echo"
)

func NewSentryErrorHandler(e *echo.Echo, c *raven.Client, l *logrus.Logger) *SentryHTTPErrorHandler {
	return &SentryHTTPErrorHandler{e, c, l}
}

type SentryHTTPErrorHandler struct {
	echo   *echo.Echo
	client *raven.Client
	logger *logrus.Logger
}

func (h *SentryHTTPErrorHandler) OnError(err error, c echo.Context) {
	h.echo.DefaultHTTPErrorHandler(err, c)
	flags := map[string]string{
		"endpoint": c.Request().RequestURI,
	}
	msg := &raven.Message{Message: err.Error(), Params: []interface{}{err}}
	rvHttp := raven.NewHttp(c.Request())
	if len(c.Request().PostForm) > 0 {
		params := make(map[string]string)
		for key, values := range c.Request().PostForm {
			params[key] = values[0]
		}
		rvHttp.Data = params
	}
	h.client.CaptureError(err, flags, msg, rvHttp)
}

func (h *SentryHTTPErrorHandler) handleError(err error, c echo.Context) {

}
