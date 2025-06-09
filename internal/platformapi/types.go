package platformapi

type Status int

const (
	Completed Status = iota
	Pending
	Failed
	Unknown
)

func (s Status) String() string {
	switch s {
	case Completed:
		return "completed"
	case Pending:
		return "pending"
	case Failed:
		return "failed"
	}

	return "unknown"
}

// TODO move to some sort of utility file in the future
func Ptr[T any](v T) *T {
	return &v
}
