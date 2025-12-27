package domain

import "context"

type Repository interface {
	Request(ctx context.Context, adID, sellerID int64) (int64, error)
	Assign(ctx context.Context, inspectionID, inspectorID int64) error
	ListAssigned(ctx context.Context, inspectorID int64) ([]Inspection, error)
	Submit(ctx context.Context, inspectionID, inspectorID int64) error
}
