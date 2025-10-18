package network_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/liaodiansm/mtg/essentials"
	"github.com/liaodiansm/mtg/network"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/mock"
)

type DialerMock struct {
	mock.Mock
}

func (d *DialerMock) Dial(network, address string) (essentials.Conn, error) {
	args := d.Called(network, address)

	return args.Get(0).(essentials.Conn), args.Error(1) //nolint: wrapcheck, forcetypeassert
}

func (d *DialerMock) DialContext(ctx context.Context, network, address string) (essentials.Conn, error) {
	args := d.Called(ctx, network, address)

	return args.Get(0).(essentials.Conn), args.Error(1) //nolint: wrapcheck, forcetypeassert
}

type HTTPServerTestSuite struct {
	httpServer *httptest.Server
}

func (suite *HTTPServerTestSuite) SetupSuite() {
	suite.httpServer = httptest.NewServer(httpbin.NewHTTPBin().Handler())
}

func (suite *HTTPServerTestSuite) TearDownSuite() {
	suite.httpServer.Close()
}

func (suite *HTTPServerTestSuite) HTTPServerAddress() string {
	return strings.TrimPrefix(suite.httpServer.URL, "http://")
}

func (suite *HTTPServerTestSuite) MakeURL(path string) string {
	return suite.httpServer.URL + path
}

func (suite *HTTPServerTestSuite) MakeHTTPClient(dialer network.Dialer) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, address) //nolint: wrapcheck
			},
		},
	}
}
