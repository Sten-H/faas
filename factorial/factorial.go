package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"math/big"
	"strconv"
	"encoding/json"
	"log"
)

type factorialResponse struct {
	FuncName   string	`json:"funcName"`
	Value	string		`json:"value"`
}

// Will give factorial of supplied n, if n is less than 0, -1 will always be returned
func factorial (n int64) string {
	if n < 0 {
		return "-1" // Unsure how this should be handled
	}
	acc := big.NewInt(1)
	var counter int64
	for counter = 1; counter <= n; counter = counter + 1 {
		acc.Mul(acc, big.NewInt(counter))
	}
	return acc.String()
}

func funcHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Responding to request")  // used to see which container responds during development
	vars := mux.Vars(r)
	nParam := vars["n"]
	n, err := strconv.Atoi(nParam)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "n must be an integer")
		return
	}
	result := factorial(int64(n))
	data := factorialResponse{
		FuncName:   "Factorial",
		Value: result,
	}
	bytes, err := json.Marshal(data)
	if err != nil {  // Unsure when this would occur
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s \n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func main() {
	fmt.Println("Starting FACTORIAL server...")
	r := mux.NewRouter()
	r.HandleFunc("/", funcHandler).Queries("n", "{n}")
	http.ListenAndServe(":8080", r)
}