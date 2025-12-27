package domain

type Status string

const (
	StatusRequested  Status = "requested"
	StatusAssigned   Status = "assigned"
	StatusInProgress Status = "in_progress"
	StatusSubmitted  Status = "submitted"
	StatusApproved   Status = "approved"
)

type Inspection struct {
	ID          int64
	AdID        int64
	SellerID    int64
	InspectorID *int64
	Status      Status
}
