package domain

type InspectionStatus string

const (
	InspectionNone       InspectionStatus = "none"
	InspectionRequested  InspectionStatus = "requested"
	InspectionInProgress InspectionStatus = "in_progress"
	InspectionDone       InspectionStatus = "done"
	InspectionCertified  InspectionStatus = "certified"
)

type AdStatus string

const (
	AdDraft      AdStatus = "draft"
	AdModeration AdStatus = "moderation"
	AdPublished  AdStatus = "published"
	AdRejected   AdStatus = "rejected"
)

type Ad struct {
	ID              int64
	SellerID        int64
	Brand           string
	Model           string
	Year            int
	Mileage         int
	Price           int
	VIN             string
	City            string
	Status          AdStatus
	InspectionState InspectionStatus
}
