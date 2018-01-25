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

// Routes to function in path with params
func gatewayRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedFunction := vars["requestedFunction"]
	// Right now I'm thinking it can just be agreed upon that port for all microservice containers is 8080
	// But perhaps this doesn't function well with swarm stuff,
	// I'm guessing the faas.name, faas.port labels were set for a reason. Haven't figured out what yet though
	response, err := http.Get("http://" + requestedFunction + ":8080" + "?" + r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Function does not exist")
		return
	}
	// Gateways send response from inner service back to client
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