package helpers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/storage"
)

func Valid(order string) bool {
	number, err := strconv.Atoi(order)
	if err != nil {
		log.Error().Err(err).Msg("")
		return false
	}
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

func SetCookie(w http.ResponseWriter, userID int, login, signingKeySTR string) {
	expirationTime := time.Now().Add(30 * time.Minute)
	signingKey := []byte(signingKeySTR)

	type Claim struct {
		UserID int    `json:"user_id"`
		Login  string `json:"login"`
		jwt.RegisteredClaims
	}

	// Create the Claims
	claims := Claim{
		userID,
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

func ErrJSONResponse(msg string, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	respErr := storage.JSONResponse{Message: msg}
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(respErr)
	if err != nil {
		log.Error().Err(err)
	}

}
