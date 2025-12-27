package application

import (
	"context"

	"autera/internal/modules/reports/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetByAdID(ctx context.Context, adID int64) (*domain.Report, error) {
	return s.repo.GetByAdID(ctx, adID)
}

func (s *Service) Dashboard(ctx context.Context) (map[string]any, error) {
	return s.repo.Dashboard(ctx)
}
