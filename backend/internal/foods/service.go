package foods

import (
	"context"
	"strings"

	"github.com/dhruvv301292/nutrichat/internal/nutrition"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, query string) ([]nutrition.Food, error) {
	normalized := strings.TrimSpace(query)
	return s.repo.Search(ctx, normalized)
}

func (s *Service) List(ctx context.Context) ([]nutrition.Food, error) {
	return s.repo.List(ctx)
}


