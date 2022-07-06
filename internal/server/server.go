package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/config"
	"github.com/yledovskikh/ya-diploma/internal/db"
	"github.com/yledovskikh/ya-diploma/internal/handlers"
	"github.com/yledovskikh/ya-diploma/internal/processing"
)

var tokenAuth *jwtauth.JWTAuth

/*func init() {
	tokenAuth = jwtauth.New("HS256", []byte("sf1o1i2rb2n2ILKJBaavkugsp23"), nil)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}*/

func Exec(ctx context.Context, wg *sync.WaitGroup) {
	// Logger

	cfg := config.GetConfig()

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	signingKey := hex.EncodeToString(b)
	log.Debug().Msg(signingKey)
	tokenAuth = jwtauth.New("HS256", []byte(signingKey), nil)
	//
	////test
	logger := httplog.NewLogger("ya-practicum", httplog.Options{
		JSON:     true,
		LogLevel: "Debug",
		Concise:  true,
	})

	d, err := db.New(cfg.DatabaseURI, ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	go processing.Exec(d, ctx, wg)
	h := handlers.New(d, signingKey)

	// Service
	r := chi.NewRouter()
	//r.Use(middleware.RequestID)
	//r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	//r.Use(middleware.Heartbeat("/ping"))

	// Protected routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Post("/api/user/orders", h.PostOrders)
		r.Get("/api/user/orders", h.GetOrders)
		//r.Get("/balance", balance.GetBalance)
	})

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.PostRegister)
		r.Post("/api/user/login", h.PostLogin)
	})
	srv := &http.Server{
		Addr:    cfg.RunAddress,
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
