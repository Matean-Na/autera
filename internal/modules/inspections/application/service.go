package application

import (
	"context"

	"autera/internal/modules/inspections/domain"
)

type Service struct{ repo domain.Repository }

func NewService(repo domain.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Request(ctx context.Context, adID, sellerID int64) (int64, error) {
	return s.repo.Request(ctx, adID, sellerID)
}

func (s *Service) Assign(ctx context.Context, inspectionID, inspectorID int64) error {
	return s.repo.Assign(ctx, inspectionID, inspectorID)
}

func (s *Service) ListAssigned(ctx context.Context, inspectorID int64) ([]domain.Inspection, error) {
	return s.repo.ListAssigned(ctx, inspectorID)
}

func (s *Service) Submit(ctx context.Context, inspectionID, inspectorID int64) error {
	return s.repo.Submit(ctx, inspectionID, inspectorID)
}
