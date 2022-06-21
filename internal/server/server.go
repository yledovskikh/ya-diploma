package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/handlers/balance"
	"github.com/yledovskikh/ya-diploma/internal/handlers/login"
	"github.com/yledovskikh/ya-diploma/internal/handlers/orders"
	"github.com/yledovskikh/ya-diploma/internal/handlers/register"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

func Exec(ctx context.Context, wg *sync.WaitGroup) {
	// Logger
	logger := httplog.NewLogger("httplog-example", httplog.Options{
		JSON: true,
	})

	// Service
	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Heartbeat("/ping"))

	// Protected routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Get("/orders", orders.GetOrders)
		r.Get("/balance", balance.GetBalance)
	})

	r.Group(func(r chi.Router) {
		r.Post("/register", register.PostRegister)
		r.Post("/login", login.PostLogin)
	})
	srv := &http.Server{
		Addr:    ":5555",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("")
		}
	}()
	log.Info().Msg("HTTP Server Started")
	<-ctx.Done()
	log.Info().Msg("HTTP Server Stopped")
	wg.Done()
}
