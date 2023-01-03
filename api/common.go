package api

// State is the state for a row.
type State string

const (
	// Active is the state for a normal row.
	Active State = "ACTIVE"
	// Deleted is the state for an removed row.
	Deleted State = "DELETED"
)
