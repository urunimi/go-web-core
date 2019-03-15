package echo

import (
	"github.com/getsentry/raven-go"
	"github.com/labstack/echo"
)

func NewSentryErrorHandler(c *raven.Client) *SentryHTTPErrorHandler {
	return &SentryHTTPErrorHandler{c}
}

type SentryHTTPErrorHandler struct {
	client *raven.Client
}

func (h *SentryHTTPErrorHandler) OnError(err error, c echo.Context) {
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
