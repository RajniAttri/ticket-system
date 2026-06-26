package models

import "time"


type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

// allowedTransitions encodes the required state machine:
//
//	open -> in_progress -> closed
//	closed -> (nothing; terminal)
var allowedTransitions = map[Status][]Status{
	StatusOpen:       {StatusInProgress},
	StatusInProgress: {StatusClosed},
}

// IsValid reports whether s is one of the known statuses. Used to reject
// garbage input like {"status":"banana"} with a 400.
func (s Status) IsValid() bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

func (s Status) CanTransitionTo(next Status) bool {
	for _, allowed := range allowedTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

type Ticket struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	OwnerID     string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
