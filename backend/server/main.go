package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chris-daniels/sportsbook-tool/odds_api"
)

type Response struct {
	Results []*odds_api.Offer `json:"results"`
}

func getOffers(w http.ResponseWriter, r *http.Request) {
	offers, err := odds_api.FetchOffers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	offersJSON, err := json.Marshal(Response{Results: offers})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return offers as JSON on response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(offersJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/offers", getOffers)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
