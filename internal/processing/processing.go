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
	storage              storage.Storage
	accrualSystemAddress string
}

func Exec(s storage.Storage, ctx context.Context, wg *sync.WaitGroup, accrualSystemAddress string) {
	defer wg.Done()
	p := Process{s, accrualSystemAddress}
	ch := make(chan int)
	var wgp sync.WaitGroup
	wgp.Add(1)
	go p.procOrders(ch, &wgp)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("exit from processing")
			//ожидаем когда полностью доработает текущий цикл procOrders
			wgp.Wait()
			return
		case <-ch:
			log.Debug().Msg("processing new part of orders")
			wgp.Add(1)
			go p.procOrders(ch, &wgp)
		}
	}
}

func (p *Process) procOrders(ch chan int, wgp *sync.WaitGroup) {
	orders, err := p.storage.GetProcOrders()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	for order, status := range orders {
		fmt.Println(order, status)
		rs := p.checkStatusOrder(order)
		//обрабатываем ответ http.StatusTooManyRequests
		//тормозим проверку статуса заказов в системе лояльности на 2сек.
		//текущий заказ мы пропускаем, но он будет обработан при следующей выборке p.storage.GetProcOrders()
		if rs == http.StatusTooManyRequests {
			time.Sleep(2 * time.Second)
		}
	}
	//Делает паузу на 5сек, прежде чем повторно будет запущена procOrders из цикла for{}, таким образом мы запускаем procOrders в цикле последовательно
	time.Sleep(5 * time.Second)
	wgp.Done()
	ch <- 1

}

func (p *Process) checkStatusOrder(o string) int {

	url := p.accrualSystemAddress + "/api/orders/" + o

	//log.Debug().Msgf("URL accrual server - %s", url)
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
