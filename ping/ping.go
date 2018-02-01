package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
	"encoding/json"
)

type pingResponse struct {
	FuncName string      `json:"funcName"`
	Value   []pingResult `json:"value"`
}

type pingResult struct {
	Ip    string	`json:"ip"`
	Rtt   string	`json:"rtt"`
}

func pingAddress(host string, resultChan chan []pingResult, errChan chan error) {
	p := fastping.NewPinger()
	ipAdress, err := net.ResolveIPAddr("ip4:icmp", host) // get Ip
	if err != nil {
		errChan <- err
		return
	}
	var result []pingResult
	p.AddIPAddr(ipAdress)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		data := pingResult{
			Ip:    addr.String(),
			Rtt:   rtt.String(),
		}
		result = append(result, data)
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	p.RunLoop()
	ticker := time.NewTicker(time.Second * 5)
	<- ticker.C
	resultChan <- result
	ticker.Stop()
	p.Stop()
	fmt.Println("pingResult finished")
}

func funcHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostAddress := vars["host"]
	resultChan, errChan := make(chan []pingResult), make(chan error)
	go pingAddress(hostAddress, resultChan, errChan)
	select {
		case result := <-resultChan:
			pingResponse := pingResponse{
				FuncName: "ping",
				Value: result,
			}
			bytes, err := json.Marshal(pingResponse)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %s \n", err)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(bytes)
			}
		case err := <-errChan:
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %s \n", err)
	}
	close(resultChan)
	close(errChan)
}

// example query: /lambda/ping?host=duckduckgo.com
func main() {
	fmt.Println("Starting PING server...")
	r := mux.NewRouter()
	r.HandleFunc("/", funcHandler).Queries("host", "{host}")
	http.ListenAndServe(":8080", r)
}