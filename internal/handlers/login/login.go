package login

import "net/http"

func PostLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login"))
}
