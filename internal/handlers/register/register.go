package register

import (
	"net/http"
)

func PostRegister(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Register"))
}
