package sitesdataservice

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	svc "github.com/mts-test-task/pkg/sitesdataservice"
	"github.com/mts-test-task/pkg/sitesdataservice/api"
)

// loggingMiddleware wraps Service and logs request information to the provided logger
type loggingMiddleware struct {
	logger log.Logger
	svc    svc.Service
}

func (s *loggingMiddleware) GetDataFromURLs(ctx context.Context, urls []string) (response []*api.SiteData, err error) {
	defer func(begin time.Time) {
		_ = s.wrap(err).Log(
			"method", "GetDataFromURLs",
			"urls", urls,
			"err", err,
			"elapsed", time.Since(begin),
		)
	}(time.Now())
	return s.svc.GetDataFromURLs(ctx, urls)
}

func (s *loggingMiddleware) wrap(err error) log.Logger {
	lvl := level.Debug
	if err != nil {
		lvl = level.Error
	}
	return lvl(s.logger)
}

// NewLoggingMiddleware ...
func NewLoggingMiddleware(logger log.Logger, svc svc.Service) svc.Service {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}
