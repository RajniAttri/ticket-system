# Ticket System (Golang)

A small backend service where a user can register, log in, create tickets, view
**only their own** tickets, and move a ticket through its status lifecycle.
Built with Go, JWT auth, bcrypt password hashing, and an in-memory store behind
a repository interface.

---

## Tech stack

| Concern | Choice |
|---------|--------|
| Language | Go 1.26 |
| Router | [chi](https://github.com/go-chi/chi) |
| Auth | JWT (`golang-jwt/jwt/v5`) + bcrypt (`golang.org/x/crypto`) |
| Storage | In-memory map behind a `store.Store` interface (swappable for SQLite) |
| Config | Environment variables via `godotenv` |

---

## Project structure

```
ticket-system/
├── cmd/server/main.go          # entry point — builds & wires everything (composition root)
├── internal/
│   ├── config/                 # loads PORT / JWT_SECRET from the environment
│   ├── httpx/                  # JSON + Error response helpers
│   ├── models/                 # User, Ticket, and the status state machine
│   ├── store/                  # repository INTERFACE + in-memory implementation
│   ├── auth/                   # bcrypt hashing + JWT generate/verify
│   ├── middleware/             # Bearer-token auth gate
│   ├── handlers/               # health, auth, ticket controllers
│   └── router/                 # route wiring (public + protected)
├── Dockerfile
├── .env.example
└── README.md
```

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Port the HTTP server listens on |
| `JWT_SECRET` | `dev-secret-change-me` | Secret used to sign/verify JWTs |

Copy `.env.example` to `.env` for local development:

```bash
cp .env.example .env
```

---

## Run locally (without Docker)

```bash
go build -o server.exe ./cmd/server
./server.exe
# → ticket-system listening on :8080
```

Health check:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

> **Windows note:** if port 8080 is stuck from a previous run, clear it with:
> ```powershell
> Get-NetTCPConnection -LocalPort 8080 -State Listen | ForEach-Object { Stop-Process -Id $_.OwningProcess -Force }
> ```

---

## Run with Docker

```bash
docker build -t ticket-system .
docker run -p 8080:8080 ticket-system
curl http://localhost:8080/health
# {"status":"ok"}
```

To pass a custom secret:

```bash
docker run -p 8080:8080 -e JWT_SECRET=my-long-random-secret ticket-system
```

---

## Run from Docker Hub (no source code / no build needed)

The image is published publicly, so anyone can pull and run it directly:

```bash
docker run -p 8080:8080 cvghnjm/ticket-system:latest
curl http://localhost:8080/health
# {"status":"ok"}
```

Docker will pull the image automatically the first time. With a custom secret:

```bash
docker run -p 8080:8080 -e JWT_SECRET=my-long-random-secret cvghnjm/ticket-system:latest
```

---

## API reference

Base URL: `http://localhost:8080`. All request/response bodies are JSON.
Protected routes require an `Authorization: Bearer <token>` header.

| Method | Endpoint | Auth | Purpose |
|--------|----------|------|---------|
| GET | `/health` | – | Health check |
| POST | `/auth/register` | – | Register a new user |
| POST | `/auth/login` | – | Log in, returns a JWT |
| POST | `/tickets` | ✅ | Create a ticket (owner = caller) |
| GET | `/tickets` | ✅ | List the caller's tickets |
| GET | `/tickets/{id}` | ✅ | Get one of the caller's tickets |
| PATCH | `/tickets/{id}/status` | ✅ | Update the status of a ticket |

### Register

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```
`201 Created` → returns the user (password hash is never included).

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```
`200 OK` → `{"token":"<JWT>"}`

### Create ticket

```bash
curl -X POST http://localhost:8080/tickets \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Login broken","description":"cannot log in"}'
```
`201 Created` → returns the ticket (`status` starts as `open`).

### List / get tickets

```bash
curl http://localhost:8080/tickets          -H "Authorization: Bearer <JWT>"
curl http://localhost:8080/tickets/<id>     -H "Authorization: Bearer <JWT>"
```

### Update status

```bash
curl -X PATCH http://localhost:8080/tickets/<id>/status \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"status":"in_progress"}'
```

---

## Status flow

```
open  ->  in_progress  ->  closed
```

- A `closed` ticket **cannot** move back to `open` or `in_progress`.
- Skipping a step (e.g. `open -> closed`) is rejected.
- An illegal transition returns `409 Conflict`; an unknown status value returns `400`.

---

## HTTP status codes used

| Code | When |
|------|------|
| 200 | Successful read/update |
| 201 | User or ticket created |
| 400 | Invalid body / invalid status value |
| 401 | Missing/invalid token, or wrong login credentials |
| 404 | Ticket not found **or** not owned by the caller |
| 409 | Duplicate email, or illegal status transition |
| 500 | Unexpected server error |

---

## Assumptions

- **In-memory storage**: data resets when the server restarts (allowed by the
  brief). The store sits behind an interface, so a persistent backend (e.g.
  SQLite) can replace it without changing handlers.
- **Ownership returns 404** (not 403) for tickets owned by another user, so the
  API does not reveal that another user's ticket id exists.
- **Password minimum length** is 6 characters; email must contain `@`.
- **JWT lifetime** is 24 hours.
- A `404` is returned for a ticket that exists but is not yours — identical to a
  ticket that does not exist — to avoid leaking ownership information.

---

## Deployment

- **Deployed URL:** _TODO — add after deploying_
- **Public health check:** _TODO — `<deployed-url>/health`_
