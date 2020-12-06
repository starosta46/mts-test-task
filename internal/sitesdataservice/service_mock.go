package sitesdataservice

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mts-test-task/pkg/sitesdataservice/api"
)

// MockService ...
type MockService struct {
	mock.Mock
}

// GetDataFromURLs ...
func (s *MockService) GetDataFromURLs(ctx context.Context, urls []string) (response []*api.SiteData, err error) {
	args := s.Called(context.Background(), urls)
	if a, ok := args.Get(0).([]*api.SiteData); ok {
		return a, args.Error(1)
	}
	return nil, args.Error(1)
}
