package sitesdataservice

import (
	"context"

	"github.com/mts-test-task/pkg/sitesdataservice/api"
)

// Service ...
type Service interface {
	GetDataFromURLs(ctx context.Context, urls []string) (response []*api.SiteData, err error)
}
