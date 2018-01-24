package main

import (
	"fmt"
	"context"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
)

// Dict where image faas.name label returns gives faas.port, perhaps this is wholly uneeded and it can
// just be decided that all microservices will run on port 8080. But I would still like to try if the faas.name
// exists in running containers I think?
var activeContainers = make(map[string]string)

// Populates activeContainers, Runs on start, most likely race condition depending on order of container creation.
// Should async poll for new labels somehow to keep updated. Docker client has ContainerDiff function.
func populateActiveContainers() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	for _, container := range containers {
		name := container.Labels["faas.name"]
		if name != "" {
			activeContainers[name] = container.Labels["faas.port"]
		}
	}
	fmt.Println("active containers populated")
}

// Test func to have a peek at containers
func listContainers(w http.ResponseWriter, r *http.Request) {
	for name, port := range activeContainers {
		fmt.Fprintf(w, "name: %s port: %s \n", name, port)
	}
}

// Checks if function specified in path exists as label name of running container and if so routes to it.
func gatewayRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedFunction := vars["requestedFunction"]
	port, ok := activeContainers[requestedFunction]
	if ok {
		response, _ := http.Get("http://" + requestedFunction + ":" + port + "?" + r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		body, _ := ioutil.ReadAll(response.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Function does not exist")
	}
}

func main() {
	fmt.Println("STARTING GATEWAY")
	r := mux.NewRouter()
	r.HandleFunc("/lambda/{requestedFunction}", gatewayRouter)
	r.HandleFunc("/containers", listContainers)
	populateActiveContainers()
	http.ListenAndServe(":80", r)
}