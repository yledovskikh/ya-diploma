package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go server.Exec(ctx, &wg)
	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-done
	cancel()
	wg.Wait()
	log.Info().Msg("shutdown server gophermart")
}
