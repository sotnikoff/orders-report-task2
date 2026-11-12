package domain

type Status string

const (
	StatusPaid      Status = "paid"
	StatusPending   Status = "pending"
	StatusCancelled Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPaid, StatusPending, StatusCancelled:
		return true
	default:
		return false
	}
}
