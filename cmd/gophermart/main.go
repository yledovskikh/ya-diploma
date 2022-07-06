package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log.Logger = log.With().Caller().Logger()
	//zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go server.Exec(ctx, &wg)
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-done
	cancel()
	wg.Wait()
	log.Info().Msg("shutdown server gophermart")
}
