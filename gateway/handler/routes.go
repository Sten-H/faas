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
	"log"
)

type RouteInfo struct {
	PathName string // path to container (ex: "faas_factorial_1")
	Port     string // Container label faas.port
	Method   string // Container label faas.method
	ID       string
}

type RouteTable struct {
	table	map[string][]RouteInfo  // Key is function function name ex "lambda/{functionName}?query=3
									// functionName is set in docker-compose label as faas.name
	lock	sync.Mutex
}

// Returns the unique path to container (used to route to it)
func getContainerPath(c types.Container) (string, error) {
	if len(c.Names) < 1 {  // Unsure when this would happen
		return "", errors.New("container has no path")
	}
	// c.Names[0] is for example "/faas_factorial_1", "/" is removed with [1:] slice
	return c.Names[0][1:], nil
}

// Returns true if container info is entered in table as RouteInfo
func (r* RouteTable) routeExists(c types.Container) bool {
	funcPath := c.Labels["faas.name"]
	routes := r.table[funcPath]
	for _, route := range routes {
		if c.ID == route.ID {
			return true
		}
	}
	return false
}

// Enter container info to into RouteTable as a RouteInfo struct
func (r* RouteTable) addRoute(c types.Container) {
	path, err := getContainerPath(c)
	if err != nil {
		log.Println(err)
		return
	}
	pathRoutes := r.table[c.Labels["faas.name"]]
	route := RouteInfo{path, c.Labels["faas.port"], c.Labels["faas.method"], c.ID,}
	pathRoutes = append(pathRoutes, route)
	r.table[c.Labels["faas.name"]] = pathRoutes
	fmt.Println("ROUTE ADDED")
}

// Checks against all containers and adds routes that do not already exist in RouteTable
func (r* RouteTable) addNewRoutes() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Println(err)
		return
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, c := range containers {
		if !r.routeExists(c) {
			r.addRoute(c)
		}
	}
}

// Removes routes which ID exists in RouteTable but not in list of containers
func (r* RouteTable) removeDeadRoutes() {
	cli, _  := client.NewEnvClient()
	containers, _ := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	containerMap := make(map[string] bool) // Will use this map as a set basically
	for _, c := range containers {
		containerMap[c.ID] = true
	}
	for pathName, paths := range r.table {
		var activePaths []RouteInfo
		for _, route := range paths {
			if containerMap[route.ID] {
				activePaths = append(activePaths, route)
			} else {
				fmt.Println("ROUTE REMOVED")
			}
		}
		r.table[pathName] = activePaths
	}
}

func New() RouteTable {
	return RouteTable{}
}

// Prunes dead routes and adds new routes
func (r *RouteTable) Update() {
	r.lock.Lock()
	r.removeDeadRoutes()
	r.addNewRoutes()
	r.lock.Unlock()
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
			case <- quit:
				ticker.Stop()
				close(quit)
				return
			}
		}
	}()
}

// Returns route info to container which will handle request
// Treats list of routes as FIFO queue as a form of load balancing
func (r *RouteTable) Get(funcPath string, method string) (RouteInfo, error) {
	r.lock.Lock()
	queue := r.table[funcPath]
	if len(queue) == 0 || queue[0].PathName == "" {
		r.lock.Unlock()
		return RouteInfo{}, errors.New("function does not exist in routing table")
	}
	route, updatedQueue := queue[0], queue[1:]
	// Check if request method matchtes function method
	if strings.ToUpper(route.Method) != strings.ToUpper(method) {
		r.lock.Unlock()
		return RouteInfo{}, errors.New("function does not exist in routing table")
	}
	updatedQueue = append(updatedQueue, route)
	r.table[funcPath] = updatedQueue
	r.lock.Unlock()
	return route, nil
}

// Initialises RouteTable by adding existing routes and scheduling updates in given interval
func (r *RouteTable) Init(updateInterval time.Duration) {
	r.table = make(map[string][]RouteInfo) // Init map
	r.addNewRoutes()
	r.ScheduleUpdates(5000)
}