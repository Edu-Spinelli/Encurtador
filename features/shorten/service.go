package shorten

import (
	"context"
	"fmt"

	"github.com/spinelli/encurtador-links/shared/encoder"
	"github.com/spinelli/encurtador-links/shared/idgen"
	"github.com/spinelli/encurtador-links/shared/metrics"
	"github.com/spinelli/encurtador-links/shared/repository"
)

type Service struct {
	baseURL string
	idGen   *idgen.Generator
	enc     *encoder.Encoder
	repo    repository.URLRepository
}

func NewService(baseURL string, idGen *idgen.Generator, enc *encoder.Encoder, repo repository.URLRepository) *Service {
	return &Service{baseURL: baseURL, idGen: idGen, enc: enc, repo: repo}
}

func (s *Service) Execute(ctx context.Context, cmd ShortenCommand) (*ShortenResponse, error) {
	id, err := s.idGen.Next(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar ID: %w", err)
	}

	shortCode, err := s.enc.Encode(id)
	if err != nil {
		return nil, fmt.Errorf("falha ao codificar ID: %w", err)
	}

	if err := s.repo.Save(shortCode, cmd.URL); err != nil {
		return nil, fmt.Errorf("falha ao salvar URL: %w", err)
	}

	metrics.UrlsShortenedTotal.Inc()

	return &ShortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", s.baseURL, shortCode),
	}, nil
}
