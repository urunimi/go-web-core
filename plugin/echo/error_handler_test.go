package echo_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/pkg/errors"

	"github.com/getsentry/raven-go"
	"github.com/urunimi/go-web-core/core"
	"github.com/urunimi/go-web-core/plugin/echo"
)

func TestNewSentryErrorHandler(t *testing.T) {
	e := core.NewEngine()
	handler := echo.NewSentryErrorHandler(raven.DefaultClient)
	u, _ := url.Parse("http://hovans.com")
	r := http.Request{Method: http.MethodGet, URL: u}
	err := errors.New("test error")

	handler.OnError(err, e.NewContext(&r, nil))
}
