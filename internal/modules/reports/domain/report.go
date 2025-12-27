package domain

type Report struct {
	ID           int64
	InspectionID int64
	TotalScore   int
	Label        string
}
