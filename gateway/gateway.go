package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
	"errors"
	"context"
	"time"
	"sync"
)

type routeInfo struct {
	name string  // Image label faas.name
	port string  // Image label faas.port
}

var routingTable = make(map[string]routeInfo)
var routeLock = sync.Mutex{}

func getFunction(funcName string) (routeInfo, error) {
	routeLock.Lock()
	data := routingTable[funcName]
	routeLock.Unlock()
	if data.name == "" {  // Zero value of name field indicates data struct is zero value as well
		return routeInfo{}, errors.New("function does not exist in routing table")
	}
	return data, nil
}

// FIXME right now it rebuilds the map entirely every run, pretty heavy handed. I do this for now as a simple
// means to delete routes that no longer exist.
func populateRoutingTable() {
	routingTable = make(map[string]routeInfo)  // Clear map
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		funcName := container.Labels["faas.name"]
		_, err := getFunction(funcName)
		if funcName != "" && err != nil { // funcName is "" on gateway
			//fmt.Printf("Putting name: %s port: %s", container.Labels["faas.name"], container.Labels["faas.port"])
			routeLock.Lock()
			routingTable[funcName] = routeInfo{ container.Labels["faas.name"], container.Labels["faas.port"]}
			routeLock.Unlock()
		}
	}
}

func scheduleRoutePopulation(msInterval time.Duration) {
	ticker := time.NewTicker(msInterval * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				populateRoutingTable()
			case <- quit:
				ticker.Stop()
				close(quit)
				return
			}
		}
	}()
}

// Routes to function in path with params
func gatewayRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedFuncName := vars["requestedFunction"]
	requestedFunc, err := getFunction(requestedFuncName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Function does not exist")
		return
	}
	url := fmt.Sprintf("http://%s:%s?%s", requestedFunc.name, requestedFunc.port, r.URL.RawQuery)
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
	r := mux.NewRouter()
	r.HandleFunc("/lambda/{requestedFunction}", gatewayRouter)
	populateRoutingTable()
	scheduleRoutePopulation(1000)  // run once per second.
	http.ListenAndServe(":80", r)
}