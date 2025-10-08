package model

type (
	// ErrNotFound
	ErrNotFound struct {
		What string
	}
)

func (e *ErrNotFound) Error() string {
	return e.What + " not found"
}
