package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strconv"
	"github.com/gorilla/mux"
	"math/big"
)

type Response struct {
	FuncName   string
	Value	string
}

// FIXME can't manage to send the big.Int in the response body, it returns as {}. Avoided by converting to string.
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
	vars := mux.Vars(r)
	nParam := vars["n"]
	n, err := strconv.Atoi(nParam)
	if err != nil {  // FIXME this doesn't catch ?n=s for example, still get 200 OK
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := factorial(int64(n))
	data := Response{
		FuncName:   "Factorial",
		Value: result,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func main() {
	fmt.Println("Starting FACTORIAL server...")
	r := mux.NewRouter()
	r.HandleFunc("/", funcHandler).Queries("n", "{n}")
	http.ListenAndServe(":8080", r)
}