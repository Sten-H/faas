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
	requestedFunc, err := routingTable.Get(requestedFuncName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Function does not exist")
		return
	}
	url := fmt.Sprintf("http://%s:%s?%s", requestedFunc.Name, requestedFunc.Port, r.URL.RawQuery)
	fmt.Println("URL: " + url)
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
	r.HandleFunc("/lambda/{requestedFunction}", gatewayRouter)
	routingTable.Populate()
	routingTable.ScheduleUpdates(10000)
	http.ListenAndServe(":80", r)
}