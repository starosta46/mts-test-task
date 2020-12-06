package httpclient

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/mts-test-task/internal/sitesdataservice"
	"github.com/mts-test-task/internal/sitesdataservice/httpserver"
	svc "github.com/mts-test-task/pkg/sitesdataservice"
	"github.com/mts-test-task/pkg/sitesdataservice/api"
	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
)

const (
	serverAddr               = "localhost:38812"
	hostAddr                 = "localhost:38812"
	maxConns                 = 512
	maxRequestBodySize       = 15 * 1024 * 1024
	serverTimeout            = 1 * time.Millisecond
	serverLaunchingWaitSleep = 1 * time.Second

	getDataFromURLsSuccess = "GetDataFromURLs success test"

	getDataFromURLsFail = "GetDataFromURLs fail test"

	serviceMethodGetDataFromURLs = "GetDataFromURLs"

	ozonURL  = "http://ozon.ru"
	wikiURL  = "https://ru.wikipedia.org"
	siteData = "html"
	fail     = "fail"
)

var (
	nilError error
)

func TestClient_GetDataFromURLsSuccess(t *testing.T) {
	urls := []string{ozonURL, wikiURL}
	response := []*api.SiteData{
		{
			URL:  wikiURL,
			Data: siteData,
		},
		{
			URL:  ozonURL,
			Data: siteData,
		},
	}
	t.Run(getDataFromURLsSuccess, func(t *testing.T) {
		serviceMock := new(sitesdataservice.MockService)
		serviceMock.On(serviceMethodGetDataFromURLs, context.Background(), urls).
			Return(response, nilError).
			Once()
		server, client := makeServerClient(serverAddr, serviceMock)
		defer func() {
			err := server.Shutdown()
			if err != nil {
				log.Printf("server shut down err: %v", err)
			}
		}()
		time.Sleep(serverLaunchingWaitSleep)
		resp, err := client.GetDataFromURLs(context.Background(), urls)
		assert.Equal(t, response, resp)
		assert.NoError(t, err, "unexpected error:", err)
	})
}

func TestClient_GetDataFromURLsFail(t *testing.T) {
	urls := []string{ozonURL, wikiURL}
	var response []*api.SiteData
	t.Run(getDataFromURLsFail, func(t *testing.T) {
		serviceMock := new(sitesdataservice.MockService)
		serviceMock.On(serviceMethodGetDataFromURLs, context.Background(), urls).
			Return(response, httperror.NewError(http.StatusBadRequest, fail, fail)).
			Once()
		server, client := makeServerClient(serverAddr, serviceMock)
		defer func() {
			err := server.Shutdown()
			if err != nil {
				log.Printf("server shut down err: %v", err)
			}
		}()
		time.Sleep(serverLaunchingWaitSleep)
		resp, err := client.GetDataFromURLs(context.Background(), urls)
		assert.Equal(t, response, resp)
		assert.Equal(t, err, httperror.NewError(http.StatusBadRequest, fail, ""))
	})
}

func makeServerClient(serverAddr string, svc svc.Service) (server *fasthttp.Server, client svc.Service) {
	client = NewPreparedClient(serverAddr, hostAddr, maxConns)
	router := httpserver.NewPreparedServer(svc)
	server = &fasthttp.Server{
		Handler:            router.Handler,
		MaxRequestBodySize: maxRequestBodySize,
		ReadTimeout:        serverTimeout,
	}
	go func() {
		err := server.ListenAndServe(serverAddr)
		if err != nil {
			log.Printf("server shut down err: %v", err)
		}
	}()

	return
}
