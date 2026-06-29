package main

import (
	"log"
	"net/http"

	"ticket-system/internal/auth"
	"ticket-system/internal/config"
	"ticket-system/internal/router"
	"ticket-system/internal/store"
)

func main() {

	cfg := config.Load()

	db := store.NewInMemoryStore()
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	handler := router.New(db, jwtManager)

	addr := ":" + cfg.Port
	log.Printf("ticket-system listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
