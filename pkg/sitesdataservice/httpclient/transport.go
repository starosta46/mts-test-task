package httpclient

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/valyala/fasthttp"

	"github.com/mts-test-task/pkg/sitesdataservice/api"
)

type errorProcessor interface {
	Encode(ctx context.Context, r *fasthttp.Response, err error)
	Decode(r *fasthttp.Response) error
}

// GetDataFromURLsClientTransport transport interface
type GetDataFromURLsClientTransport interface {
	EncodeRequest(ctx context.Context, r *fasthttp.Request, urls []string) (err error)
	DecodeResponse(ctx context.Context, r *fasthttp.Response) (response []*api.SiteData, err error)
}

type getDataFromURLsClientTransport struct {
	errorProcessor errorProcessor
	pathTemplate   string
	method         string
}

// EncodeRequest method for encoding requests on client side
func (g *getDataFromURLsClientTransport) EncodeRequest(ctx context.Context, r *fasthttp.Request, urls []string) (err error) {
	r.Header.SetMethod(g.method)
	r.SetRequestURI(g.pathTemplate)
	r.Header.Set("Content-Type", "application/json")
	return json.NewEncoder(r.BodyWriter()).Encode(urls)
}

// DecodeResponse method for decoding response on client side
func (g *getDataFromURLsClientTransport) DecodeResponse(ctx context.Context, r *fasthttp.Response) (response []*api.SiteData, err error) {
	if r.StatusCode() != http.StatusOK {
		err = g.errorProcessor.Decode(r)
		return
	}
	err = json.Unmarshal(r.Body(), &response)
	return
}

// NewGetDataFromURLsClientTransport the transport creator for http requests
func NewGetDataFromURLsClientTransport(
	errorProcessor errorProcessor,
	pathTemplate string,
	method string,
) GetDataFromURLsClientTransport {
	return &getDataFromURLsClientTransport{
		errorProcessor: errorProcessor,
		pathTemplate:   pathTemplate,
		method:         method,
	}
}
