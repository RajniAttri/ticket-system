package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ticket-system/internal/httpx"
	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
)

type TicketHandler struct {
	store store.TicketStore
}

// NewTicketHandler wires the handler with its store dependency.
func  NewTicketHandler(s store.TicketStore) *TicketHandler {
	return &TicketHandler{store: s}
}

// createTicketRequest is the POST /tickets body.
type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// updateStatusRequest is the PATCH /tickets/{id}/status body.
type updateStatusRequest struct {
	Status models.Status `json:"status"`
}

// Create makes a new ticket owned by the authenticated user, in status "open".
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var body createTicketRequest
	if err := decodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		httpx.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	now := time.Now()
	ticket := &models.Ticket{
		ID:          uuid.NewString(),
		Title:       strings.TrimSpace(body.Title),
		Description: strings.TrimSpace(body.Description),
		Status:      models.StatusOpen, // every ticket starts open
		OwnerID:     userID,            // ownership stamped from the token, never the body
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.store.CreateTicket(ticket); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create ticket")
		return
	}

	httpx.JSON(w, http.StatusCreated, ticket)
}

// List returns every ticket owned by the authenticated user (and only those).
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	tickets, err := h.store.ListTicketsByOwner(userID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list tickets")
		return
	}

	httpx.JSON(w, http.StatusOK, tickets)
}

// Get returns a single ticket by id, but only if the caller owns it.
func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// chi.URLParam reads the {id} path segment — like req.params.id in Express.
	id := chi.URLParam(r, "id")

	ticket, err := h.loadOwnedTicket(id, userID)
	if err != nil {
		writeTicketLookupError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, ticket)
}

// UpdateStatus moves a ticket along the state machine, enforcing ownership and
// the open -> in_progress -> closed rules (closed is terminal).
func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id := chi.URLParam(r, "id")

	var body updateStatusRequest
	if err := decodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// 1. Reject unknown status values (e.g. "banana") with 400.
	if !body.Status.IsValid() {
		httpx.Error(w, http.StatusBadRequest, "status must be one of: open, in_progress, closed")
		return
	}

	// 2. Load the ticket and enforce ownership (404 if missing or not yours).
	ticket, err := h.loadOwnedTicket(id, userID)
	if err != nil {
		writeTicketLookupError(w, err)
		return
	}

	// 3. No-op if the status is unchanged — treat as a successful idempotent update.
	if ticket.Status == body.Status {
		httpx.JSON(w, http.StatusOK, ticket)
		return
	}

	// 4. Enforce the state machine. An illegal move (e.g. closed -> open, or
	//    skipping open -> closed) is a conflict with the resource's current
	//    state, which is exactly what 409 Conflict means.
	if !ticket.Status.CanTransitionTo(body.Status) {
		httpx.Error(w, http.StatusConflict,
			"illegal status transition from "+string(ticket.Status)+" to "+string(body.Status))
		return
	}

	ticket.Status = body.Status
	ticket.UpdatedAt = time.Now()

	if err := h.store.UpdateTicket(ticket); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not update ticket")
		return
	}

	httpx.JSON(w, http.StatusOK, ticket)
}

// --- shared ticket helpers ---

func (h *TicketHandler) loadOwnedTicket(id, userID string) (*models.Ticket, error) {
	ticket, err := h.store.GetTicketByID(id)
	if err != nil {
		return nil, err 
	}
	if ticket.OwnerID != userID {
		return nil, store.ErrNotFound
	}
	return ticket, nil
}

// writeTicketLookupError maps a lookup error to the right HTTP status.
func writeTicketLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, "ticket not found")
		return
	}
	httpx.Error(w, http.StatusInternalServerError, "could not load ticket")
}
