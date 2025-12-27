package application

import (
	"context"

	"autera/internal/modules/ads/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

type CreateAdInput struct {
	SellerID int64  `json:"seller_id"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	Mileage  int    `json:"mileage"`
	Price    int    `json:"price"`
	VIN      string `json:"vin"`
	City     string `json:"city"`
}

func (s *Service) Create(ctx context.Context, in CreateAdInput) (int64, error) {
	ad := &domain.Ad{
		SellerID:        in.SellerID,
		Brand:           in.Brand,
		Model:           in.Model,
		Year:            in.Year,
		Mileage:         in.Mileage,
		Price:           in.Price,
		VIN:             in.VIN,
		City:            in.City,
		Status:          domain.AdDraft,
		InspectionState: domain.InspectionNone,
	}
	return s.repo.Create(ctx, ad)
}

func (s *Service) Get(ctx context.Context, id int64) (*domain.Ad, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, f domain.ListFilter) ([]domain.Ad, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) SubmitToModeration(ctx context.Context, adID, sellerID int64) error {
	return s.repo.SubmitToModeration(ctx, adID, sellerID)
}

func (s *Service) Moderate(ctx context.Context, adID int64, decision string) error {
	return s.repo.Moderate(ctx, adID, decision)
}
