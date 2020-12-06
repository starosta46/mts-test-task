package httpserver

import (
	"context"
	"net/http"

	"github.com/buaazp/fasthttprouter"
	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
	"github.com/valyala/fasthttp"

	"github.com/mts-test-task/pkg/sitesdataservice/api"
)

type service interface {
	GetDataFromURLs(ctx context.Context, request []string) (response []*api.SiteData, err error)
}

type getDataFromURLsServer struct {
	transport      GetDataFromURLsTransport
	service        service
	errorProcessor httperror.ErrorProcessor
}

// ServeHTTP implements http.Handler.
func (g *getDataFromURLsServer) ServeHTTP(ctx *fasthttp.RequestCtx) {
	request, err := g.transport.DecodeRequest(ctx, &ctx.Request)
	if err != nil {
		g.errorProcessor.Encode(ctx, &ctx.Response, err)
		return
	}

	response, err := g.service.GetDataFromURLs(ctx, request)
	if err != nil {
		g.errorProcessor.Encode(ctx, &ctx.Response, err)
		return
	}

	if err := g.transport.EncodeResponse(ctx, &ctx.Response, response); err != nil {
		g.errorProcessor.Encode(ctx, &ctx.Response, err)
		return
	}
}

// NewGetDataFromURLsServer the server creator
func NewGetDataFromURLsServer(transport GetDataFromURLsTransport, service service, errorProcessor httperror.ErrorProcessor) fasthttp.RequestHandler {
	ls := getDataFromURLsServer{
		transport:      transport,
		service:        service,
		errorProcessor: errorProcessor,
	}
	return ls.ServeHTTP
}

// NewPreparedServer factory for server api handler
func NewPreparedServer(svc service) *fasthttprouter.Router {
	errorProcessor := httperror.NewErrorProcessor(http.StatusInternalServerError, "Внутряння ошибка сервиса")

	getDataFromURLsTransport := NewGetDataFromURLsTransport(httperror.NewError)

	return MakeFastHTTPRouter(
		[]*HandlerSettings{
			{
				Path:    URIPathGetDataFromURLs,
				Method:  HTTPMethodGetDataFromURLs,
				Handler: NewGetDataFromURLsServer(getDataFromURLsTransport, svc, errorProcessor),
			},
		},
	)
}
