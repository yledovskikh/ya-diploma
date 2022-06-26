package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	//"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
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

func errJSONResponse(msg string, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	respErr := storage.JSONResponse{Message: msg}
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(respErr)
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
		errJSONResponse(err.Error(), http.StatusBadRequest, w)
		return
	}

	err = s.storage.NewUser(u)

	if err != nil {
		log.Error().Err(err).Msg("")
		status, msg := storageErrToStatus(err)
		errJSONResponse(msg, status, w)
		return
	}
	s.setCookie(w, u.Login)
	response := storage.JSONResponse{Message: "User registered"}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error().Err(err)
	}
}

func (s *Server) setCookie(w http.ResponseWriter, login string) {
	expirationTime := time.Now().Add(30 * time.Minute)
	signingKey := []byte(s.signingKey)

	type Claim struct {
		Login string `json:"login"`
		jwt.RegisteredClaims
	}

	// Create the Claims
	claims := Claim{
		login,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "ya-practicum",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signingKey)
	//fmt.Printf("%v %v", ss, err)

	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		log.Error().Err(err).Msg("")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   ss,
		Expires: expirationTime,
	})

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
		errJSONResponse(err.Error(), http.StatusBadRequest, w)
		return
	}

	err = s.storage.CheckUser(u)

	if err != nil {
		status, msg := storageErrToStatus(err)
		errJSONResponse(msg, status, w)
		return
	}
	s.setCookie(w, u.Login)
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
	_, claims, _ := jwtauth.FromContext(r.Context())

	body, err := io.ReadAll(r.Body)
	// обрабатываем ошибку
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	order, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	check := valid(order)
	login := fmt.Sprintf("%v", claims["login"])
	if !check {
		log.Debug().Msgf("User %s tried to load not valid order %d", login, order)
		msg := fmt.Sprintf("Order %d is not valid", order)
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return
	}
	err = s.storage.SetOrder(login, order)
	if err != nil {
		status, msg := storageErrToStatus(err)
		http.Error(w, msg, status)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("Order %d registered by user %s", order, claims["login"])))
}

func valid(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

func storageErrToStatus(err error) (int, string) {
	//	ErrBadRequest             = errors.New(`HTTP 400 Bad Request`)
	//	ErrUnauthorized           = errors.New(`HTTP 401 Unauthorized`)
	//	ErrInternalServerError    = errors.New(`HTTP 500 Internal Server Error`)
	//	ErrLoginAlreadyExist      = errors.New(`HTTP 409 Login Already Exists`)
	//	ErrUserAlreadyLoadedOrder = errors.New(`HTTP 200 You Have Already Uploaded The Order`)
	//	ErrOrderLoadedAnotherUser = errors.New(`HTTP 409 The Order Has Already Been Uploaded By Another User`)
	switch err {
	case storage.ErrBadRequest:
		return http.StatusBadRequest, storage.ErrBadRequest.Error()
	case storage.ErrUnauthorized:
		return http.StatusUnauthorized, storage.ErrUnauthorized.Error()
	case storage.ErrLoginAlreadyExist:
		return http.StatusConflict, storage.ErrLoginAlreadyExist.Error()
	case storage.ErrUserAlreadyLoadedOrder:
		return http.StatusOK, storage.ErrUserAlreadyLoadedOrder.Error()
	case storage.ErrOrderLoadedAnotherUser:
		return http.StatusConflict, storage.ErrOrderLoadedAnotherUser.Error()
	default:
		return http.StatusInternalServerError, storage.ErrInternalServerError.Error()
	}
}
