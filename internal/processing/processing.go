package processing

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/storage"
)

type Process struct {
	storage storage.Storage
}

func Exec(s storage.Storage, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	//TODO изучить вопрос и возможно нужно что-то изменить с каналом
	c := make(chan int)
	p := Process{s}
	close(c)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("exit from processing")
			return
		case <-c:
			log.Debug().Msg("processing new part of orders")
			p.procOrders()
		}
	}
}

func (p *Process) procOrders() {
	orders, err := p.storage.GetProcOrders()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	for order, status := range orders {
		fmt.Println(order, status)
		rs := p.checkStatusOrder(order)
		if rs == http.StatusTooManyRequests {
			time.Sleep(5 * time.Second)
		}
	}
	time.Sleep(5 * time.Second)
	//c <- 0
}

func (p *Process) checkStatusOrder(o string) int {

	//TODO get endpoint accrual system from config
	url := "http://localhost:8080/api/orders/" + o

	resp, err := http.Get(url)
	log.Debug().Msgf("http get %s", url)
	if err != nil {
		log.Error().Err(err).Msg("")
		return http.StatusInternalServerError
	}
	defer resp.Body.Close()

	log.Debug().Msgf("response code is %d", resp.StatusCode)

	if resp.StatusCode == http.StatusTooManyRequests {
		log.Error().Msgf("Response code for order %s from Accrual System is %d", o, http.StatusTooManyRequests)
		return http.StatusTooManyRequests
	}

	if resp.StatusCode == http.StatusOK {
		var order storage.OrderAccrual
		err = json.NewDecoder(resp.Body).Decode(&order)
		if err != nil {
			log.Error().Err(err).Msg("")
			return 0
		}
		err = p.storage.UpdateStatusOrder(order)
		if err != nil {
			return 0
		}
	}
	return 0

}
