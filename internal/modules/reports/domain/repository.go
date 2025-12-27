package domain

import "context"

type Repository interface {
	GetByAdID(ctx context.Context, adID int64) (*Report, error)
	Dashboard(ctx context.Context) (map[string]any, error)
}
