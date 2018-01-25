package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
	"net/url"
	"encoding/json"
)

type pingResponse struct {
	FuncName string
	Values   []Ping
}

type Ping struct {
	Ip    string
	Rtt   string
	Error string
}
 // FIXME as of right now "http://" has to be included in addr for a valid response
func pingAddress(addr string, resultChan chan []Ping) {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", addr)  // get Ip
	if err != nil {
		fmt.Println(err)
		resultChan <- []Ping{ {Ip: "", Rtt: "", Error: err.Error()} }
	}
	result := []Ping{}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		data := Ping{
			Ip:    addr.String(),
			Rtt:   rtt.String(),
			Error: "",
		}
		result = append(result, data)
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	p.OnIdle = func() {
		resultChan <- result
		fmt.Println("Ping finished")
	}
	// FIXME p.RunLoop should be used to ping multiple times
	// https://godoc.org/github.com/tatsushid/go-fastping#Pinger.RunLoop
	err = p.Run()
	if err != nil {
		resultChan <- []Ping{ {Ip: "", Rtt: "", Error: err.Error()} }
	}
}

func funcHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strAddress := vars["address"]
	u, err := url.Parse(strAddress)
	if err != nil {  // this doesn't trigger on things like strAdress = 5 which would have been nice
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %s \n", err)
	}
	fmt.Printf("HOST: %s", u.Host)
	resultChan := make(chan []Ping)
	go pingAddress(u.Host, resultChan)
	result := <- resultChan  // To my understanding HandleFunc creates a new goroutine
							 // for each request so blocking should be ok
	pingResponse := pingResponse{
		FuncName: "ping",
		Values: result,
	}
	bytes, err := json.Marshal(pingResponse)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	//fmt.Fprintln(w, result)
}

// example query: /lambda/ping?address=http://www.duckduckgo.com
func main() {
	fmt.Println("Starting PING server...")
	r := mux.NewRouter()
	r.HandleFunc("/", funcHandler).Queries("address", "{address}")
	http.ListenAndServe(":8080", r)
}