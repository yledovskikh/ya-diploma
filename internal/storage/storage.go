package storage

import (
	"errors"
	"net/http"
	"time"
)

type Storage interface {
	NewUser(User) (int, error)
	CheckUser(User) (int, error)
	SetOrder(int, string) error
	GetOrders(int) ([]Order, error)
	GetBalance(int) (Balance, error)
	GetProcOrders() (map[string]string, error)
	UpdateStatusOrder(OrderAccrual) error
	PingDB() error
	Close()
}

var (
	ErrBadRequest             = errors.New(`HTTP 400 Bad Request`)
	ErrUnauthorized           = errors.New(`HTTP 401 Unauthorized`)
	ErrInternalServerError    = errors.New(`HTTP 500 Internal Server Error`)
	ErrLoginAlreadyExist      = errors.New(`HTTP 409 Login Already Exists`)
	ErrOrderLoadedAnotherUser = errors.New(`HTTP 409 The Order Has Already Been Uploaded By Another User`)
	ErrUserAlreadyLoadedOrder = errors.New(`HTTP 200 You Have Already Uploaded The Order`)
)

type User struct {
	ID       int
	Login    string
	Password string
	CreateAt time.Time
	UpdateAt time.Time
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

//            "number": "9278923470",
//            "status": "PROCESSED",
//            "accrual": 500,
//            "uploaded_at": "2020-12-10T15:15:45+03:00"

type Order struct {
	ID       string  `json:"number"`
	Status   string  `json:"status"`
	Accrual  float32 `json:"accrual"`
	CreateAt string  `json:"uploaded_at"`
}

type OrderAccrual struct {
	ID       string  `json:"order"`
	Status   string  `json:"status"`
	Accrual  float32 `json:"accrual"`
	CreateAt string  `json:"uploaded_at"`
}

type JSONResponse struct {
	Message string
}

func StorageErrToStatus(err error) (int, string) {
	//	ErrBadRequest             = errors.New(`HTTP 400 Bad Request`)
	//	ErrUnauthorized           = errors.New(`HTTP 401 Unauthorized`)
	//	ErrInternalServerError    = errors.New(`HTTP 500 Internal Server Error`)
	//	ErrLoginAlreadyExist      = errors.New(`HTTP 409 Login Already Exists`)
	//	ErrUserAlreadyLoadedOrder = errors.New(`HTTP 200 You Have Already Uploaded The Order`)
	//	ErrOrderLoadedAnotherUser = errors.New(`HTTP 409 The Order Has Already Been Uploaded By Another User`)
	switch err {
	case ErrBadRequest:
		return http.StatusBadRequest, ErrBadRequest.Error()
	case ErrUnauthorized:
		return http.StatusUnauthorized, ErrUnauthorized.Error()
	case ErrLoginAlreadyExist:
		return http.StatusConflict, ErrLoginAlreadyExist.Error()
	case ErrUserAlreadyLoadedOrder:
		return http.StatusOK, ErrUserAlreadyLoadedOrder.Error()
	case ErrOrderLoadedAnotherUser:
		return http.StatusConflict, ErrOrderLoadedAnotherUser.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServerError.Error()
	}
}
