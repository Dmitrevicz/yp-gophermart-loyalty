// Package accrual contains methods to communicate with accrual service.
// Implements AccrualService interface.
package accrual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/model"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/client"
	"github.com/Dmitrevicz/yp-gophermart-loyalty/internal/util/sync"
)

const (
	StatusRegistered = "REGISTERED" // заказ зарегистрирован, но начисление не рассчитано
	StatusInvalid    = "INVALID"    // заказ не принят к расчёту, и вознаграждение не будет начислено
	StatusProcessing = "PROCESSING" // расчёт начисления в процессе
	StatusProcessed  = "PROCESSED"  // расчёт начисления окончен
	StatusOrderNew   = "NEW"
)

var (
	pathGetOrderAccrual string = "/api/orders/"
)

const DefaultMaxReq = 32

// AccrualService implements AccrualService interface.
type AccrualService struct {
	client    *http.Client
	semaphore *sync.Semaphore
}

func New(addr string) *AccrualService {
	pathGetOrderAccrual = addr + pathGetOrderAccrual

	return &AccrualService{
		client:    client.NewClientDefault(),
		semaphore: sync.NewSemaphore(DefaultMaxReq),
	}
}

// Order - получение информации о расчёте начислений баллов лояльности.
//
// GET {accrual_service}/api/orders/{number}
func (a *AccrualService) Order(id string) (accrual model.AccrualOrder, err error) {
	a.semaphore.Acquire()
	defer a.semaphore.Release()

	url := pathGetOrderAccrual + id
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return accrual, fmt.Errorf("error preparing request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		// return nil, model.NewRetriableError(fmt.Errorf("error while doing the request: %w", err))
		return accrual, fmt.Errorf("error while doing the request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return accrual, fmt.Errorf("error reading response bytes: %w", err)
	}

	if len(body) > 0 {
		if err = json.Unmarshal(body, &accrual); err != nil {
			return accrual, fmt.Errorf("error decoding response body: %w", err)
		}
	}

	// TODO: handle other expected codes
	// Возможные коды ответа:
	//	200 - успешная обработка запроса
	//	204 - заказ не зарегистрирован в системе расчета
	//	429 - превышено количество запросов к сервису
	//	500 - внутренняя ошибка сервера
	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected response status code: %s, body: %s", resp.Status, string(body))
		// if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		// 	return nil, model.NewRetriableError(err)
		// }
		return accrual, err
	}

	return accrual, nil
}
