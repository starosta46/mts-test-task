package builder

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mts-test-task/pkg/sitesdataservice/api"
)

// SitesDataMongoObjects build necessary objects for requests to mongodb
type SitesDataMongoObjects interface {
	GetSitesDataFilter(createDate int64) (filter bson.M)
	AddSitesData(sitesData []*api.SiteData, createTime int64) (data []interface{})
	SortOptions(fieldName string, sortType int) (findOptions *options.FindOptions)
}

type sitesDataMongoObjects struct {
	createDateNameField string
	urlsNameField       string
	dataNameFiled       string
}

func (s *sitesDataMongoObjects) GetSitesDataFilter(createDate int64) (filter bson.M) {
	return bson.M{
		s.createDateNameField: bson.M{
			"$gt": createDate,
		},
	}
}

func (s *sitesDataMongoObjects) AddSitesData(sitesData []*api.SiteData, createTime int64) (data []interface{}) {
	data = make([]interface{}, len(sitesData))
	for i := 0; i < len(sitesData); i++ {
		data[i] = bson.D{
			{Key: s.createDateNameField, Value: createTime},
			{Key: s.urlsNameField, Value: sitesData[i].URL},
			{Key: s.dataNameFiled, Value: sitesData[i].Data},
		}
	}

	return
}

func (s *sitesDataMongoObjects) SortOptions(fieldName string, sortType int) (findOptions *options.FindOptions) {
	return options.Find().SetSort(bson.D{{Key: fieldName, Value: sortType}})
}

// NewSitesDataMongoObjects ...
func NewSitesDataMongoObjects(
	createDateNameField string,
	urlsNameField string,
	dataNameFiled string,
) SitesDataMongoObjects {
	return &sitesDataMongoObjects{
		createDateNameField: createDateNameField,
		urlsNameField:       urlsNameField,
		dataNameFiled:       dataNameFiled,
	}
}
