package redirect

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spinelli/encurtador-links/shared/metrics"
	"github.com/spinelli/encurtador-links/shared/repository"
)

var ErrNotFound = errors.New("URL não encontrada")

const cacheTTL = 24 * time.Hour

type Service struct {
	repo  repository.URLRepository
	cache *redis.Client
}

func NewService(repo repository.URLRepository, cache *redis.Client) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) Execute(ctx context.Context, query RedirectQuery) (string, error) {
	cached, err := s.cache.Get(ctx, "cache:"+query.ShortCode).Result()
	if err == nil {
		metrics.CacheHitsTotal.Inc()
		metrics.UrlsRedirectedTotal.Inc()
		return cached, nil
	}

	metrics.CacheMissesTotal.Inc()

	url, err := s.repo.FindByShortCode(query.ShortCode)
	if err != nil {
		return "", ErrNotFound
	}

	s.cache.Set(ctx, "cache:"+url.ShortCode, url.LongURL, cacheTTL)
	metrics.UrlsRedirectedTotal.Inc()

	return url.LongURL, nil
}
