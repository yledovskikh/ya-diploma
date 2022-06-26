package storage

import (
	"errors"
	"time"
)

type Storage interface {
	NewUser(User) error
	CheckUser(User) error
	SetOrder(string, int) error
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
	Login    string
	Password string
	CreateAt time.Time
	UpdateAt time.Time
}
type JSONResponse struct {
	Message string
}
