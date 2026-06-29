package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticket-system/internal/auth"
	"ticket-system/internal/router"
	"ticket-system/internal/store"
)

type testEnv struct {
	t       *testing.T
	handler http.Handler
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	s := store.NewInMemoryStore()
	jwtManager := auth.NewJWTManager("test-secret")
	return &testEnv{t: t, handler: router.New(s, jwtManager)}
}

func (e *testEnv) do(method, path, token string, body any) *httptest.ResponseRecorder {
	e.t.Helper()

	var reader *bytes.Reader
	switch v := body.(type) {
	case nil:
		reader = bytes.NewReader(nil)
	case string:
		reader = bytes.NewReader([]byte(v))
	default:
		raw, _ := json.Marshal(v)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, want, rec.Body.String())
	}
}

// registerAndLogin sets up a user and returns a usable token.
func (e *testEnv) registerAndLogin(email string) string {
	e.t.Helper()
	creds := map[string]string{"email": email, "password": "password123"}
	if rec := e.do(http.MethodPost, "/auth/register", "", creds); rec.Code != http.StatusCreated {
		e.t.Fatalf("register = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	rec := e.do(http.MethodPost, "/auth/login", "", creds)
	assertStatus(e.t, rec, http.StatusOK)
	var resp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	return resp.Token
}

func (e *testEnv) createTicket(token, title string) string {
	e.t.Helper()
	rec := e.do(http.MethodPost, "/tickets", token, map[string]string{"title": title})
	assertStatus(e.t, rec, http.StatusCreated)
	var ticket struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &ticket)
	return ticket.ID
}

func TestHealth(t *testing.T) {
	e := newTestEnv(t)
	assertStatus(t, e.do(http.MethodGet, "/health", "", nil), http.StatusOK)
}

func TestRegister(t *testing.T) {
	e := newTestEnv(t)
	creds := map[string]string{"email": "user@example.com", "password": "password123"}

	// Success.
	assertStatus(t, e.do(http.MethodPost, "/auth/register", "", creds), http.StatusCreated)
	// Duplicate email.
	assertStatus(t, e.do(http.MethodPost, "/auth/register", "", creds), http.StatusConflict)
	// Invalid input.
	assertStatus(t, e.do(http.MethodPost, "/auth/register", "", map[string]string{"email": "bad", "password": "x"}), http.StatusBadRequest)
}

func TestLogin(t *testing.T) {
	e := newTestEnv(t)
	e.registerAndLogin("login@example.com") // also asserts the success path

	// Wrong password.
	rec := e.do(http.MethodPost, "/auth/login", "", map[string]string{"email": "login@example.com", "password": "wrong"})
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestProtectedRoutesRequireAuth(t *testing.T) {
	e := newTestEnv(t)
	assertStatus(t, e.do(http.MethodGet, "/tickets", "", nil), http.StatusUnauthorized)
	assertStatus(t, e.do(http.MethodGet, "/tickets", "bad-token", nil), http.StatusUnauthorized)
}

func TestCreateTicket(t *testing.T) {
	e := newTestEnv(t)
	token := e.registerAndLogin("owner@example.com")

	// Success: new tickets start "open".
	rec := e.do(http.MethodPost, "/tickets", token, map[string]string{"title": "My ticket"})
	assertStatus(t, rec, http.StatusCreated)
	var ticket map[string]any
	json.Unmarshal(rec.Body.Bytes(), &ticket)
	if ticket["status"] != "open" {
		t.Fatalf("status = %v, want open", ticket["status"])
	}

	// Missing title.
	assertStatus(t, e.do(http.MethodPost, "/tickets", token, map[string]string{"title": "  "}), http.StatusBadRequest)
}

func TestListAndGetTicket(t *testing.T) {
	e := newTestEnv(t)
	token := e.registerAndLogin("owner@example.com")
	id := e.createTicket(token, "fetch me")

	// List returns the owner's tickets.
	rec := e.do(http.MethodGet, "/tickets", token, nil)
	assertStatus(t, rec, http.StatusOK)
	var list []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}

	// Get by id.
	assertStatus(t, e.do(http.MethodGet, "/tickets/"+id, token, nil), http.StatusOK)
	// Unknown id.
	assertStatus(t, e.do(http.MethodGet, "/tickets/nope", token, nil), http.StatusNotFound)
}

func TestTicketOwnershipIsolation(t *testing.T) {
	e := newTestEnv(t)
	alice := e.registerAndLogin("alice@example.com")
	bob := e.registerAndLogin("bob@example.com")
	id := e.createTicket(alice, "alice secret")

	// Bob can't see Alice's ticket — looks like a 404, not a 403.
	assertStatus(t, e.do(http.MethodGet, "/tickets/"+id, bob, nil), http.StatusNotFound)
}

func TestUpdateStatus(t *testing.T) {
	e := newTestEnv(t)
	token := e.registerAndLogin("owner@example.com")
	id := e.createTicket(token, "lifecycle")

	patch := func(status string) *httptest.ResponseRecorder {
		return e.do(http.MethodPatch, "/tickets/"+id+"/status", token, map[string]string{"status": status})
	}

	// Legal transition: open -> in_progress.
	assertStatus(t, patch("in_progress"), http.StatusOK)
	// Invalid status value.
	assertStatus(t, patch("banana"), http.StatusBadRequest)
	// Illegal transition: in_progress -> open.
	assertStatus(t, patch("open"), http.StatusConflict)
}
