package main

import (
	"fmt"
	//"context"
	"net/http"
	"github.com/gorilla/mux"
	//"github.com/docker/docker/api/types"
	//"github.com/docker/docker/client"
	"io/ioutil"
)

// Checks if function specified in path exists as label name of running container and if so routes to it.
func gatewayRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedFunction := vars["requestedFunction"]
	response, err := http.Get("http://" + requestedFunction + ":8080" + "?" + r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Function does not exist")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	body, _ := ioutil.ReadAll(response.Body)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func main() {
	fmt.Println("STARTING GATEWAY")
	r := mux.NewRouter()
	r.HandleFunc("/lambda/{requestedFunction}", gatewayRouter)
	http.ListenAndServe(":80", r)
}