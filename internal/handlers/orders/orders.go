package orders

import (
	"fmt"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

func GetOrders(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("Orders"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Write([]byte(fmt.Sprintf("protected area - Orders. hi %v", claims["user_id"])))
}
