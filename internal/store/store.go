package store

import (
	"errors"

	"ticket-system/internal/models"
)


var (
	ErrNotFound    = errors.New("not found")
	ErrEmailExists = errors.New("email already registered")
)


type UserStore interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
}


type TicketStore interface {
	CreateTicket(ticket *models.Ticket) error
	GetTicketByID(id string) (*models.Ticket, error)
	ListTicketsByOwner(ownerID string) ([]*models.Ticket, error)
	UpdateTicket(ticket *models.Ticket) error
}

type Store interface {
	UserStore
	TicketStore
}
