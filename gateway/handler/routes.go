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

type FunctionInfo struct {
	lock *sync.RWMutex  // Should never be accessed directly, use getLock()
	routes []RouteInfo
	currentIndex int
}

func (f* FunctionInfo) incIndex() {
	f.currentIndex = (f.currentIndex + 1) % len(f.routes)
}

func (f* FunctionInfo) getLock() *sync.RWMutex {
	if f.lock == nil {
		f.lock = &sync.RWMutex{}
	}
	return f.lock
}

func (f* FunctionInfo) getNextRoute() (RouteInfo, error) {
	if len(f.routes) <= f.currentIndex {
		return RouteInfo{}, errors.New("function does not exist in routing table")
	}
	route := f.routes[f.currentIndex]
	if strings.ToUpper(route.Method) != strings.ToUpper(route.Method) {
		return RouteInfo{}, errors.New("function does not exist in routing table")
	}
	return route, nil
}

func (f* FunctionInfo) addRoute(c types.Container) {
	path, err := getContainerPath(c)
	if err != nil {
		log.Println(err)
		return
	}
	route := RouteInfo{path, c.Labels["faas.port"], c.Labels["faas.method"], c.ID,}
	f.routes = append(f.routes, route)
	fmt.Println("ROUTE ADDED")
}

func (f* FunctionInfo) setRoutes(routes []RouteInfo) {
	f.routes = routes
	f.currentIndex = f.currentIndex % len(f.routes)
}

type RouteTable struct {
	table	map[string]FunctionInfo  // Key is function function name ex "lambda/{functionName}?query=3		// functionName is set in docker-compose label as faas.name
}

// Returns the unique path to container (used to route to it)
func getContainerPath(c types.Container) (string, error) {
	if len(c.Names) < 1 {  // Unsure when this would happen
		return "", errors.New("container has no path")
	}
	// c.Names[0] is for example "/faas_factorial_1", "/" is removed with [1:] slice
	return c.Names[0][1:], nil
}

// Checks against all containers and adds routes that do not already exist in RouteTable
func (r* RouteTable) addNewRoutes(containers []types.Container) {
	routeMap := make(map[string] bool)  // used as set
	for _, funcInfo := range r.table {
		funcInfo.lock.RLock()
		for _, route := range funcInfo.routes {
			routeMap[route.ID] = true
		}
		funcInfo.lock.RUnlock()
	}
	for _, c := range containers {
		if !routeMap[c.ID] {
			funcPath := c.Labels["faas.name"]
			funcInfo := r.table[funcPath]
			funcInfo.addRoute(c)
			funcInfo.getLock().Lock()
			r.table[funcPath] = funcInfo
			funcInfo.getLock().Unlock()
		}
	}
}

// Removes routes which ID exists in RouteTable but not in list of containers
func (r* RouteTable) removeDeadRoutes(containers []types.Container) {
	containerMap := make(map[string] bool) // used as set
	for _, c := range containers {
		containerMap[c.ID] = true
	}
	for pathName, funcInfo := range r.table {
		lock := funcInfo.getLock()
		lock.RLock()
		var activeRoutes []RouteInfo
		for _, route := range funcInfo.routes {
			if containerMap[route.ID] {
				activeRoutes = append(activeRoutes, route)
			} else {
				fmt.Println("ROUTE REMOVED")
			}
		}
		funcInfo.setRoutes(activeRoutes)
		lock.RUnlock()
		lock.Lock()
		r.table[pathName] = funcInfo
		lock.Unlock()
	}
}

func New() RouteTable {
	return RouteTable{}
}

// Removes dead routes and adds new routes
func (r *RouteTable) Update() {
	cli, err  := client.NewEnvClient()
	if err != nil {
		log.Println(err)
		return
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Println(err)
		return
	}
	r.addNewRoutes(containers)
	r.removeDeadRoutes(containers)
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

// Returns route which will handle request
func (r *RouteTable) Get(funcPath string, method string) (RouteInfo, error) {
	funcInfo := r.table[funcPath]
	lock := funcInfo.getLock()
	lock.RLock()
	route, err := funcInfo.getNextRoute()
	if err != nil {
		lock.RUnlock()
		return RouteInfo{}, err
	}
	funcInfo.incIndex()
	lock.RUnlock()
	lock.Lock()
	r.table[funcPath] = funcInfo
	lock.Unlock()
	return route, nil
}

// Initialises RouteTable by adding existing routes and scheduling updates in given interval
func (r *RouteTable) Init(updateInterval time.Duration) {
	r.table = make(map[string]FunctionInfo) // Init map
	r.Update()
	r.ScheduleUpdates(updateInterval)
}