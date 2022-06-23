package storage

import "errors"

type Storage interface {
	NewUser(User) error
	CheckUser(string, string) (bool, error)
	PingDB() error
	Close()
}

var (
	ErrBadRequest          = errors.New(`HTTP 400 Bad Request`)
	ErrInternalServerError = errors.New(`HTTP 500 Internal Server Error`)
	ErrLoginAlreadyExist   = errors.New(`HTTP 409 Login Already Exists`)
)

type User struct {
	Login    string
	Password string
}
type JSONResponse struct {
	Message string
}
