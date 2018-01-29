package handler

import (
	"errors"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sync"
	"time"
	"fmt"
	"strings"
)

type RouteInfo struct {
	Name string // Image label faas.Name
	Port string // Image label faas.port
	Method string  // Image label faas.method
}

type RouteTable struct {
	table	map[string][]RouteInfo  // Key is function functionPath ex "lambda/{functionPath}?query=3
									// functionPath is set in docker-compose label as faas.path
	lock	sync.Mutex
}

// Returns the unique path to container (used to route to it)
func getContainerPath(c types.Container) (string, error) {
	if len(c.Names) < 1 {  // Unsure when this would happen
		return "", errors.New("container has no path")
	}
	// c.Names[0] is for example "/faas_factorial_1", "/" is removed with slice
	return c.Names[0][1:], nil
}

func New() RouteTable {
	return RouteTable{}
}

// Populates RouteTable map with all docker container paths that can be found by image name
func (r *RouteTable) Populate() {
	r.lock.Lock()
	r.table = make(map[string][]RouteInfo) // Clear map
	r.lock.Unlock()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}
	r.lock.Lock()
	for _, container := range containers {
		funcPath := container.Labels["faas.path"]
		containerPath, err := getContainerPath(container)
		if err != nil || container.Labels["faas.Name"] == "gateway" {
			continue
		}
		r.table[funcPath] = append(r.table[funcPath],
			RouteInfo{
				containerPath,
				container.Labels["faas.port"],
				container.Labels["faas.method"],
				})
	}
	r.lock.Unlock()
}

// FIXME right now is mocked. It will just call to rebuild the entire table.
func (r *RouteTable) Update() {
	r.Populate()
}

// Calls RouteTable.Update in the interval given
func (r *RouteTable) ScheduleUpdates(msInterval time.Duration) {
	ticker := time.NewTicker(msInterval * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				r.Update()
				fmt.Println("Routes updated")  // FIXME for some overview during development
			case <- quit:
				ticker.Stop()
				close(quit)
				return
			}
		}
	}()
}

// Returns route info to container which will handle request
// Right now this uses a very naive form of "load balancing" where the list of available containers is
// treated as a queue and the returned path is taken from fron of queue and then put in back of queue.
func (r *RouteTable) Get(imageName string, method string) (RouteInfo, error) {
	r.lock.Lock()
	queue := r.table[imageName]
	if len(queue) == 0 || queue[0].Name == "" {
		return RouteInfo{}, errors.New("function does not exist in routing table")
	}
	// Check if request method matchtes function method
	route := queue[0]
	if strings.ToUpper(route.Method) != strings.ToUpper(method) {
		return RouteInfo{}, errors.New("function does not exist in routing table")
	}
	queue = queue[1:]
	queue = append(queue, route)
	r.table[imageName] = queue
	r.lock.Unlock()
	return route, nil
}