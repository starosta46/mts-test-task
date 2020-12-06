package converter

import "github.com/mts-test-task/internal/storages/mongodb/models"

//SitesData convert side data to map
type SitesData interface {
	SitesDataToMap(sitesData []*models.SiteData) (sitesDataMap map[string]*models.SiteData)
}

type sitesData struct{}

func (s *sitesData) SitesDataToMap(sitesData []*models.SiteData) (sitesDataMap map[string]*models.SiteData) {
	sitesDataMap = make(map[string]*models.SiteData)
	for i := 0; i < len(sitesData); i++ {
		sitesDataMap[sitesData[i].URL] = sitesData[i]
	}

	return
}

//NewSitesData ...
func NewSitesData() SitesData {
	return &sitesData{}
}
