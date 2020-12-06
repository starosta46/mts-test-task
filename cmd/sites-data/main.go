package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kelseyhightower/envconfig"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/mts-test-task/internal/converter"
	"github.com/mts-test-task/internal/sites"
	"github.com/mts-test-task/internal/sitesdataservice"
	"github.com/mts-test-task/internal/sitesdataservice/httpserver"
	"github.com/mts-test-task/internal/storages/mongodb/builder"
	"github.com/mts-test-task/internal/storages/mongodb/wrapper"
	"github.com/mts-test-task/internal/validator"
	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
)

// Название полей в mongodb
const (
	createDateNameField = "create_date"
	urlsNameField       = "url"
	dataNameFiled       = "data"
)

// Направление сортировки
const (
	sortAsc = 1
)

type configuration struct {
	// Настройки сервера
	Port                 string        `envconfig:"PORT" default:"8080"`
	MaxRequestBodySize   int           `envconfig:"MAX_REQUEST_BODY_SIZE" default:"10485760"` // 10 MB
	MaxSimultaneousConns int           `envconfig:"MAX_SIM_CONNS" default:"100"`
	ServerTimeout        time.Duration `envconfig:"SERVER_TIMEOUT" default:"10000ms"`

	// Отображение логов успешных запросов
	Debug bool `envconfig:"DEBUG" default:"true"`

	// Максимальное число урлов для обработки
	MaxURLsCount int `envconfig:"MAX_URLS_COUNT" default:"20"`

	// Таймаут запроса к сайтам
	SitesClientTimeout time.Duration `envconfig:"SITES_CLIENT_TIMEOUT" default:"500ms"`

	// Настройки mongodb
	SitesDataMongoCollection string        `envconfig:"SITES_DATA_MONGO_COLLECTION" default:"sites"`
	MongoAddr                []string      `envconfig:"MONGO_ADDR" default:"127.0.0.1:27017"`
	MongoDBName              string        `envconfig:"MONGO_DB_NAME" default:"sites"`
	MongoDBUser              string        `envconfig:"MONGO_DB_USER" default:"root"`
	MongoDBPass              string        `envconfig:"MONGO_DB_PASS" default:"rootpassword"`
	SitesDataMongoTimeout    time.Duration `envconfig:"SITES_DATA_MONGO_TIMEOUT" default:"1000ms"`
}

func main() {
	// logger
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// processing configuration
	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		_ = level.Error(logger).Log("msg", "failed to load configuration", "err", err)
		os.Exit(1)
	}

	if !cfg.Debug {
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	inputValidator := validator.NewInput(cfg.MaxURLsCount, httperror.NewError)

	sitesClient := sites.NewClient(http.Client{Timeout: cfg.SitesClientTimeout})

	// mongo storage
	ctxTimeout, cancel := context.WithTimeout(context.Background(), cfg.SitesDataMongoTimeout)
	defer cancel()

	sitesDataMongoClientOptions := options.ClientOptions{
		Auth: &options.Credential{
			Username: cfg.MongoDBUser,
			Password: cfg.MongoDBPass,
		},
		Hosts: cfg.MongoAddr,
	}
	sitesDataMongoClient, err := mongo.Connect(ctxTimeout, &sitesDataMongoClientOptions)
	if err != nil {
		_ = level.Error(logger).Log("msg", "failed initialize connection to storage", "err", err)
		os.Exit(1)
	}
	defer func() {
		err = sitesDataMongoClient.Disconnect(ctxTimeout)
		if err != nil {
			_ = level.Error(logger).Log("msg", "error disconnecting from mongodb", "err", err)
		}
	}()
	if err != nil || sitesDataMongoClient == nil {
		_ = level.Error(logger).Log("msg", "error connecting to mongodb", "err", err)
		os.Exit(1)
	}
	if err = sitesDataMongoClient.Ping(ctxTimeout, readpref.Primary()); err != nil {
		_ = level.Error(logger).Log("msg", "couldn't ping database, exiting", "err", err)
		os.Exit(1)
	}

	sitesDataMongoClientDB := sitesDataMongoClient.Database(cfg.MongoDBName)
	sitesDataMongoWrapper := wrapper.NewSitesDataWrapper(
		sitesDataMongoClientDB,
		cfg.SitesDataMongoCollection,
		cfg.SitesDataMongoTimeout,
	)

	svc := sitesdataservice.NewService(
		inputValidator,
		httperror.NewError,
		sitesClient,
		builder.NewSitesDataMongoObjects(
			createDateNameField,
			urlsNameField,
			dataNameFiled),
		sitesDataMongoWrapper,
		logger,
		converter.NewSitesData(),
		createDateNameField,
		sortAsc,
	)
	// Добавляем метрики к сервису
	svc = sitesdataservice.NewLoggingMiddleware(logger, svc)

	router := httpserver.NewPreparedServer(svc)

	// Устанавливаем таймаут сервера и ошибку
	handlerWithTimeout := fasthttp.TimeoutHandler(
		router.Handler,
		cfg.ServerTimeout,
		"Обработка запроса превысила установленный таймаут",
	)

	fasthttpServer := &fasthttp.Server{
		Handler: handlerWithTimeout,
		// Устанавливаем максимальное число одновременных запросов
		Concurrency:        cfg.MaxSimultaneousConns,
		MaxRequestBodySize: cfg.MaxRequestBodySize,
	}

	go func() {
		_ = level.Info(logger).Log("msg", "starting http server", "port", cfg.Port)
		if err := fasthttpServer.ListenAndServe(":" + cfg.Port); err != nil {
			_ = level.Error(logger).Log("msg", "server run failure", "err", err)
			os.Exit(1)
		}
	}()
	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer func(sig os.Signal) {
		_ = level.Info(logger).Log("msg", "received signal, exiting", "signal", sig)

		if err := fasthttpServer.Shutdown(); err != nil {
			_ = level.Error(logger).Log("msg", "server shutdown failure", "err", err)
		}

		_ = level.Info(logger).Log("msg", "server stopped")
	}(<-c)
}
