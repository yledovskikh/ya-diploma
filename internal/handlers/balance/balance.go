package balance

import (
	"fmt"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("Orders"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Write([]byte(fmt.Sprintf("protected area - Balance. hi %v", claims["user_id"])))
}
