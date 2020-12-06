package wrapper

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mts-test-task/internal/storages/mongodb/models"
)

const (
	errorFind   = "Find() err: %s"
	errorDecode = "Decode() err: %s"
	errorInsert = "InsertMany() err: %s"
)

// SitesDataWrapper ...
type SitesDataWrapper interface {
	GetSitesData(ctx context.Context, filter bson.M, sortOptions *options.FindOptions) (sitesData []*models.SiteData, err error)
	AddSitesData(ctx context.Context, data []interface{}) (err error)
}

type sitesDataWrapper struct {
	database       *mongo.Database
	collectionName string
	timeout        time.Duration
}

func (s *sitesDataWrapper) GetSitesData(ctx context.Context, filter bson.M, sortOptions *options.FindOptions) (sitesData []*models.SiteData, err error) {
	sitesData = make([]*models.SiteData, 0)

	ctxTimeOut, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	res, err := s.database.Collection(s.collectionName).Find(ctxTimeOut, filter, sortOptions)
	if err != nil {
		err = fmt.Errorf(errorFind, err)
		return
	}

	if err = res.All(ctx, &sitesData); err != nil {
		err = fmt.Errorf(errorDecode, err)
		return
	}

	return
}

func (s *sitesDataWrapper) AddSitesData(ctx context.Context, data []interface{}) (err error) {
	ctxTimeOut, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err = s.database.Collection(s.collectionName).InsertMany(ctxTimeOut, data)
	if err != nil {
		err = fmt.Errorf(errorInsert, err)
	}

	return
}

// NewSitesDataWrapper ...
func NewSitesDataWrapper(
	database *mongo.Database,
	collectionName string,
	timeout time.Duration,
) SitesDataWrapper {
	return &sitesDataWrapper{
		database:       database,
		collectionName: collectionName,
		timeout:        timeout,
	}
}
