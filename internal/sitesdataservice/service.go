package sitesdataservice

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"

	"github.com/mts-test-task/internal/storages/mongodb/models"
	svc "github.com/mts-test-task/pkg/sitesdataservice"
	"github.com/mts-test-task/pkg/sitesdataservice/api"
	"github.com/mts-test-task/pkg/sitesdataservice/httperror"
)

type inputValidator interface {
	CheckURLs(urls []string) (err error)
}

type sitesClient interface {
	GetData(ctx context.Context, url string) (data string, err error)
}

type sitesDataMongoObjectsBuilder interface {
	GetSitesDataFilter(createDate int64) (filter bson.M)
	AddSitesData(sitesData []*api.SiteData, createTime int64) (data []interface{})
	SortOptions(fieldName string, sortType int) (findOptions *options.FindOptions)
}

type sitesDataMongoWrapper interface {
	GetSitesData(ctx context.Context, filter bson.M, sortOptions *options.FindOptions) (sitesData []*models.SiteData, err error)
	AddSitesData(ctx context.Context, data []interface{}) (err error)
}

type sitesDataConverter interface {
	SitesDataToMap(sitesData []*models.SiteData) (sitesDataMap map[string]*models.SiteData)
}

type service struct {
	inputValidator               inputValidator
	errorCreator                 httperror.ErrorCreator
	sitesClient                  sitesClient
	sitesDataMongoObjectsBuilder sitesDataMongoObjectsBuilder
	sitesDataMongoWrapper        sitesDataMongoWrapper
	logger                       log.Logger
	sitesDataConverter           sitesDataConverter
	createDataFieldName          string
	sortAsc                      int
}

func (s *service) GetDataFromURLs(ctx context.Context, urls []string) (response []*api.SiteData, err error) {
	// Проверяем входные данные
	err = s.inputValidator.CheckURLs(urls)
	if err != nil {
		return
	}

	// Данные для запроса в монгу
	filter := s.sitesDataMongoObjectsBuilder.GetSitesDataFilter(time.Now().Add(-time.Minute).Unix())
	sortOptions := s.sitesDataMongoObjectsBuilder.SortOptions(s.createDataFieldName, s.sortAsc)
	// Получаем результаты предыдущих запросов из монги
	storedSitesData, err := s.sitesDataMongoWrapper.GetSitesData(ctx, filter, sortOptions)
	if err != nil {
		// Будем считать, что монгу использем как кэш, поэтому ошибку будем логировать,
		// но не будем прерывать выполнение программы
		_ = level.Error(s.logger).Log("msg", "Failed to get data from sites data mongo:", "err", err)
		err = nil
	}

	// Конвертируем результаты предыдущих запросов в мапу, чтобы можно было быстрее их получать
	storedSitesDataMap := s.sitesDataConverter.SitesDataToMap(storedSitesData)

	// Создаем errgroup для работы с горутинами
	group, ctx := errgroup.WithContext(ctx)

	response = make([]*api.SiteData, len(urls))

	for i := range urls {
		iteration := i
		group.Go(func() error {
			if storedSiteData, ok := storedSitesDataMap[urls[iteration]]; ok {
				response[iteration] = &api.SiteData{URL: urls[iteration], Data: storedSiteData.Data[:1]}
			} else {
				siteData, err := s.sitesClient.GetData(ctx, urls[iteration])
				if err != nil {
					return s.errorCreator(
						http.StatusBadGateway,
						fmt.Sprintf("Не удалось получить данные от %s", urls[iteration]),
						fmt.Sprintf("failed to get data from %s: %s", urls[iteration], err),
					)
				}
				response[iteration] = &api.SiteData{URL: urls[iteration], Data: siteData}
			}
			return nil
		})

	}

	err = group.Wait()
	if err != nil {
		return
	}

	// Подготавливаем данные для сохранения в монгу
	siteDataToStore := s.sitesDataMongoObjectsBuilder.AddSitesData(response, time.Now().Unix())
	err = s.sitesDataMongoWrapper.AddSitesData(ctx, siteDataToStore)
	if err != nil {
		_ = level.Error(s.logger).Log("msg", "Failed to put data to sites data mongo:", "err", err)
		err = nil
	}

	return
}

// NewService ...
func NewService(
	inputValidator inputValidator,
	errorCreator httperror.ErrorCreator,
	sitesClient sitesClient,
	sitesDataMongoObjectsBuilder sitesDataMongoObjectsBuilder,
	sitesDataMongoWrapper sitesDataMongoWrapper,
	logger log.Logger,
	sitesDataConverter sitesDataConverter,
	createDataFieldName string,
	sortAsc int,
) svc.Service {
	return &service{
		inputValidator:               inputValidator,
		errorCreator:                 errorCreator,
		sitesClient:                  sitesClient,
		sitesDataMongoObjectsBuilder: sitesDataMongoObjectsBuilder,
		sitesDataMongoWrapper:        sitesDataMongoWrapper,
		logger:                       logger,
		sitesDataConverter:           sitesDataConverter,
		createDataFieldName:          createDataFieldName,
		sortAsc:                      sortAsc,
	}
}
