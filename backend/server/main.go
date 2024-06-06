package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chris-daniels/sportsbook-tool/database"
	"github.com/chris-daniels/sportsbook-tool/odds_api"
)

type Response struct {
	Results []*odds_api.Offer `json:"results"`
}

func postBet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST /bets")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Content-Type", "application/json")

	// only allow POST requests
	if r.Method != http.MethodPost {
		fmt.Println("Method not allowed")
		// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		w.WriteHeader(http.StatusOK)
		return
	}
	// Parse post request body for event id
	offerBody := &odds_api.Offer{}
	err := json.NewDecoder(r.Body).Decode(&offerBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.InsertOffer(offerBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
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
	http.HandleFunc("/bets", postBet)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
