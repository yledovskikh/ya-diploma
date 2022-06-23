package handlers

import (
	"encoding/json"
	"errors"

	//"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/storage"
)

//var (
//	ErrBadRequest     = errors.New("invalid value")
//	ErrNotFound       = errors.New("metric not found")
//	ErrNotImplemented = errors.New("unknown metric type")
//)

type Server struct {
	storage storage.Storage
}

func New(s storage.Storage) *Server {
	return &Server{storage: s}
}

func errJSONResponse(err error, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	respErr := storage.JSONResponse{Message: err.Error()}
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(respErr)
	if err != nil {
		log.Error().Err(err)
	}

}

func (s Server) PostRegister(w http.ResponseWriter, r *http.Request) {

	//- `200` — пользователь успешно зарегистрирован и аутентифицирован;
	//- `400` — неверный формат запроса;
	//- `409` — логин уже занят;
	//- `500` — внутренняя ошибка сервера.

	var u storage.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Error().Err(err)
		errJSONResponse(err, http.StatusBadRequest, w)
		return
	}

	err = s.storage.NewUser(u)

	if err != nil {
		status := http.StatusConflict
		if errors.Is(err, storage.ErrInternalServerError) {
			status = http.StatusInternalServerError
		}
		errJSONResponse(err, status, w)
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("")
	}
	response := storage.JSONResponse{Message: "User registered"}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error().Err(err)
	}

	//w.Write([]byte("Register"))
}

func (s Server) PostLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login"))
}
