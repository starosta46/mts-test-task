package httpclient

import (
	"context"
	"net/http"

	"github.com/valyala/fasthttp"

	svc "github.com/mts-test-task/pkg/sitesdataservice"
	"github.com/mts-test-task/pkg/sitesdataservice/api"
	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
)

type client struct {
	cli *fasthttp.HostClient

	transportGetDataFromURLs GetDataFromURLsClientTransport
}

// GetDataFromURLs ...
func (s *client) GetDataFromURLs(ctx context.Context, request []string) (response []*api.SiteData, err error) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	if err = s.transportGetDataFromURLs.EncodeRequest(ctx, req, request); err != nil {
		return
	}

	err = s.cli.Do(req, res)
	if err != nil {
		return
	}
	return s.transportGetDataFromURLs.DecodeResponse(ctx, res)
}

// NewClient the client creator
func NewClient(
	cli *fasthttp.HostClient,

	transportGetDataFromURLs GetDataFromURLsClientTransport,
) svc.Service {
	return &client{
		cli: cli,

		transportGetDataFromURLs: transportGetDataFromURLs,
	}
}

// NewPreparedClient create and set up http client
func NewPreparedClient(
	serverURL string,
	serverHost string,
	maxConns int,
) svc.Service {
	errorProcessor := httperror.NewErrorProcessor(http.StatusInternalServerError, "Внутряння ошибка сервиса")
	transportGetDataFromURLs := NewGetDataFromURLsClientTransport(
		errorProcessor,
		MethodHTTP+serverURL+URIPathClientGetDataFromURLs,
		HTTPMethodClientGetDataFromURLs,
	)

	return NewClient(
		&fasthttp.HostClient{
			Addr:     serverHost,
			MaxConns: maxConns,
		},

		transportGetDataFromURLs,
	)
}
