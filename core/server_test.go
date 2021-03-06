package core_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/urunimi/gorest/core"
)

type ServerTestSuite struct {
	suite.Suite
	app    *TestApp
	server core.Server
}

func TestNewServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (sts *ServerTestSuite) SetupTest() {
	_ = os.Setenv("server_env", "test")
	sts.app = &TestApp{}
	sts.server = core.NewServer(sts.app)
}

func (sts *ServerTestSuite) TestServer_NewServer() {
	a := assert.Assertions{}
	a.NotNil(sts.server)
}

func (sts *ServerTestSuite) TestServer_Init() {
	a := assert.Assertions{}
	_ = sts.server.Init()
	a.Equal(sts.app.inited, true)
}

func (sts *ServerTestSuite) TestServer_Start_Exit() {
	a := assert.Assertions{}
	_ = sts.server.Init()
	var wg sync.WaitGroup
	wg.Add(2)
	go func(sts *ServerTestSuite) {
		sts.server.Start()
		wg.Done()
	}(sts)
	go func(sts *ServerTestSuite) {
		for sts.app.registered == false {
			time.Sleep(time.Millisecond)
		}
		a.Equal(true, sts.app.registered)
		sts.server.Exit(nil)
		wg.Done()
	}(sts)
	wg.Wait()
	a.Equal(true, sts.app.cleaned)
}

type TestApp struct {
	inited, registered, cleaned bool
}

func (ts *TestApp) Init() error {
	ts.inited = true
	return nil
}

func (ts *TestApp) RegisterRoute(serviceDriver *core.Engine) {
	ts.registered = true
}

func (ts *TestApp) Clean() error {
	ts.inited = false
	ts.registered = false
	ts.cleaned = true
	return nil
}
