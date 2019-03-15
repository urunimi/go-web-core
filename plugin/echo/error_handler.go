package echo

import (
	"github.com/Sirupsen/logrus"
	"github.com/getsentry/raven-go"
	"github.com/labstack/echo"
)

func SetErrorHandlerForSentry(e *echo.Echo, c *raven.Client, l *logrus.Logger) {
	chh := SentryHTTPErrorHandler{e, c, l}
	e.HTTPErrorHandler = chh.handleError
}

type SentryHTTPErrorHandler struct {
	echo   *echo.Echo
	client *raven.Client
	logger *logrus.Logger
}

func (h *SentryHTTPErrorHandler) handleError(err error, c echo.Context) {
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
	//h.logger.Warn("%v\n%v\nerr: %s", c.Request().URL, rvHttp.Data, err.Error())
	h.client.CaptureError(err, flags, msg, rvHttp)
}
