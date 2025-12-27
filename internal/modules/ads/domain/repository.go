package domain

import "context"

type ListFilter struct {
	VerifiedOnly  *bool
	Brand         string
	City          string
	YearFrom      *int
	YearTo        *int
	PriceFrom     *int
	PriceTo       *int
	MileageFrom   *int
	MileageTo     *int
	Inspection    string
	Limit, Offset int
}

type Repository interface {
	Create(ctx context.Context, ad *Ad) (int64, error)
	Get(ctx context.Context, id int64) (*Ad, error)
	List(ctx context.Context, f ListFilter) ([]Ad, int64, error)

	SubmitToModeration(ctx context.Context, adID, sellerID int64) error
	Moderate(ctx context.Context, adID int64, decision string) error
}
