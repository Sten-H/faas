package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"github.com/sten-H/faas/gateway/handler"
)

var routingTable handler.RouteTable

// Routes to function in path with params
func gatewayRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedFuncName := vars["requestedFunction"]
	requestedFunc, err := routingTable.Get(requestedFuncName, r.Method)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Function does not exist")
		return
	}
	url := fmt.Sprintf("http://%s:%s?%s", requestedFunc.PathName, requestedFunc.Port, r.URL.RawQuery)
	fmt.Println("Routing to " + url)
	response, err := http.Get(url)
	if err != nil {  // This error should be extremely unlikely but still possible I think
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Function could not be reached")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode) // Copy status code
	body, _ := ioutil.ReadAll(response.Body)
	w.Write(body)
}

func main() {
	fmt.Println("STARTING GATEWAY")
	routingTable = handler.New()
	r := mux.NewRouter()
	r.HandleFunc("/lambda/{requestedFunction}", gatewayRouter).Methods("GET", "POST", "PUT", "DELETE")
	routingTable.Init(5000)
	http.ListenAndServe(":80", r)
}