package ping

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/tatsushid/go-fastping"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	fmt.Println("Starting PING server...")
	r := mux.NewRouter()
	r.HandleFunc("/", pingHandler).Queries("address", "{address}")
	http.ListenAndServe(":8080", r)
}