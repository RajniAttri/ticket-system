package store

import (
	"sync"

	"ticket-system/internal/models"
)


type InMemoryStore struct {
	mu           sync.RWMutex
	usersByID    map[string]*models.User
	usersByEmail map[string]*models.User
	tickets      map[string]*models.Ticket
}


var _ Store = (*InMemoryStore)(nil)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		usersByID:    make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		tickets:      make(map[string]*models.Ticket),
	}
}

// ------- USER IMPLEMENTATION -------

func (s *InMemoryStore) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[user.Email]; exists {
		return ErrEmailExists
	}

	s.usersByID[user.ID] = user
	s.usersByEmail[user.Email] = user
	return nil
}

func (s *InMemoryStore) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

func (s *InMemoryStore) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return user, nil
}

// --- TicketStore implementation ---

// CreateTicket stores a new ticket.
func (s *InMemoryStore) CreateTicket(ticket *models.Ticket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tickets[ticket.ID] = ticket
	return nil
}

func (s *InMemoryStore) GetTicketByID(id string) (*models.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, ok := s.tickets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return ticket, nil
}

func (s *InMemoryStore) ListTicketsByOwner(ownerID string) ([]*models.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Ticket, 0)
	for _, ticket := range s.tickets {
		if ticket.OwnerID == ownerID {
			result = append(result, ticket)
		}
	}
	return result, nil
}
 
func (s *InMemoryStore) UpdateTicket(ticket *models.Ticket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tickets[ticket.ID]; !ok {
		return ErrNotFound
	}
	s.tickets[ticket.ID] = ticket
	return nil
}
