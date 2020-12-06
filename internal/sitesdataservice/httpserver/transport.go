package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"

	"github.com/mts-test-task/pkg/sitesdataservice/api"
	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
)

// GetDataFromURLsTransport transport interface
type GetDataFromURLsTransport interface {
	DecodeRequest(ctx context.Context, r *fasthttp.Request) (urls []string, err error)
	EncodeResponse(ctx context.Context, r *fasthttp.Response, response []*api.SiteData) (err error)
}

type getDataFromURLsTransport struct {
	errorCreator httperror.ErrorCreator
}

// DecodeRequest method for decoding requests on server side
func (g *getDataFromURLsTransport) DecodeRequest(ctx context.Context, r *fasthttp.Request) (urls []string, err error) {
	if err = json.Unmarshal(r.Body(), &urls); err != nil {
		return urls, g.errorCreator(
			http.StatusBadRequest,
			"Не удалось обработать запрос",
			fmt.Sprintf("failed to decode JSON request: %v", err),
		)
	}
	return
}

// EncodeResponse method for encoding response on server side
func (g *getDataFromURLsTransport) EncodeResponse(ctx context.Context, r *fasthttp.Response, response []*api.SiteData) (err error) {
	r.Header.Set("Content-Type", "application/json")
	if err = json.NewEncoder(r.BodyWriter()).Encode(&response); err != nil {
		return g.errorCreator(
			http.StatusInternalServerError,
			"Не удалось обработать ответ",
			fmt.Sprintf("failed to encode JSON response: %s", err),
		)
	}
	return
}

// NewGetDataFromURLsTransport the transport creator for http requests
func NewGetDataFromURLsTransport(errorCreator httperror.ErrorCreator) GetDataFromURLsTransport {
	return &getDataFromURLsTransport{
		errorCreator: errorCreator,
	}
}
