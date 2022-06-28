package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	//"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/helpers"
	"github.com/yledovskikh/ya-diploma/internal/storage"
)

//var (
//	ErrBadRequest     = errors.New("invalid value")
//	ErrNotFound       = errors.New("metric not found")
//	ErrNotImplemented = errors.New("unknown metric type")
//)

type Server struct {
	storage    storage.Storage
	signingKey string
}

func New(s storage.Storage, signingKey string) *Server {
	return &Server{storage: s, signingKey: signingKey}
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
		helpers.ErrJSONResponse(err.Error(), http.StatusBadRequest, w)
		return
	}

	u.ID, err = s.storage.NewUser(u)
	if err != nil {
		log.Error().Err(err).Msg("")
		status, msg := storage.StorageErrToStatus(err)
		helpers.ErrJSONResponse(msg, status, w)
		return
	}
	helpers.SetCookie(w, u.ID, u.Login, s.signingKey)
	response := storage.JSONResponse{Message: "User registered"}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error().Err(err)
	}
}

func (s Server) PostLogin(w http.ResponseWriter, r *http.Request) {
	//- `200` — пользователь успешно аутентифицирован;
	//- `400` — неверный формат запроса;
	//- `401` — неверная пара логин/пароль;
	//- `500` — внутренняя ошибка сервера.

	var u storage.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Error().Err(err).Msg("")
		helpers.ErrJSONResponse(err.Error(), http.StatusBadRequest, w)
		return
	}

	u.ID, err = s.storage.CheckUser(u)

	if err != nil {
		status, msg := storage.StorageErrToStatus(err)
		helpers.ErrJSONResponse(msg, status, w)
		return
	}
	helpers.SetCookie(w, u.ID, u.Login, s.signingKey)
	response := storage.JSONResponse{Message: "User logged"}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error().Err(err)
	}

}

func (s Server) PostOrders(w http.ResponseWriter, r *http.Request) {
	//- `200` — номер заказа уже был загружен этим пользователем;
	//- `202` — новый номер заказа принят в обработку;
	//- `400` — неверный формат запроса;
	//- `401` — пользователь не аутентифицирован;
	//- `409` — номер заказа уже был загружен другим пользователем;
	//- `422` — неверный формат номера заказа;
	//- `500` — внутренняя ошибка сервера.

	body, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orderNumber, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	check := helpers.Valid(orderNumber)
	_, claims, _ := jwtauth.FromContext(r.Context())
	login := fmt.Sprintf("%v", claims["login"])

	if !check {
		log.Debug().Msgf("User %s tried to load not valid order %d", login, orderNumber)
		msg := fmt.Sprintf("Order %d is not valid", orderNumber)
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return
	}
	userID, err := strconv.Atoi(fmt.Sprintf("%v", claims["user_id"]))
	if err != nil {
		log.Error().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err = s.storage.SetOrder(userID, orderNumber)
	if err != nil {
		status, msg := storage.StorageErrToStatus(err)
		http.Error(w, msg, status)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_, err = w.Write([]byte(fmt.Sprintf("Order %d registered by user %s", orderNumber, claims["login"])))
	if err != nil {
		log.Error().Err(err)
	}
}

func (s *Server) GetOrders(w http.ResponseWriter, r *http.Request) {
	//
	//	- `200` — успешная обработка запроса.
	//
	//		Формат ответа:
	//
	//	```
	//    200 OK HTTP/1.1
	//    Content-Type: application/json
	//    ...
	//
	//    [
	//    	{
	//            "number": "9278923470",
	//            "status": "PROCESSED",
	//            "accrual": 500,
	//            "uploaded_at": "2020-12-10T15:15:45+03:00"
	//        },
	//        {
	//            "number": "12345678903",
	//            "status": "PROCESSING",
	//            "uploaded_at": "2020-12-10T15:12:01+03:00"
	//        },
	//        {
	//            "number": "346436439",
	//            "status": "INVALID",
	//            "uploaded_at": "2020-12-09T16:09:53+03:00"
	//        }
	//    ]
	//    ```
	//
	//	- `204` — нет данных для ответа.
	//	- `401` — пользователь не авторизован.
	//	- `500` — внутренняя ошибка сервера.

	_, claims, _ := jwtauth.FromContext(r.Context())
	userID, err := strconv.Atoi(fmt.Sprintf("%v", claims["user_id"]))
	if err != nil {
		log.Error().Err(err).Msg("")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	orders, err := s.storage.GetOrders(userID)
	if err != nil {
		status, msg := storage.StorageErrToStatus(err)
		http.Error(w, msg, status)
		return
	}
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

}
