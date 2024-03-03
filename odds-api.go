package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
)

var API_KEY = ""

func FetchOffers() ([]*Offer, error) {
	// Read API key from file
	apiKey, err := os.ReadFile("TOKEN.txt")
	if err != nil {
		return nil, err
	}
	API_KEY = string(apiKey)

	aggOffers := []*Offer{}
	// iterate over all sports and get offers
	for _, keyMarket := range KEYS_MARKETS {
		offers, err := getOffersForSport(keyMarket[0], keyMarket[1])
		if err != nil {
			return nil, err
		}
		aggOffers = append(aggOffers, offers...)
	}

	// Sort best offers by outlier score
	sort.Slice(aggOffers, func(i, j int) bool {
		return aggOffers[i].OutlierScore > aggOffers[j].OutlierScore
	})

	// filter for offers with desired bookmaker, price, and number of competitors
	filteredOffers := []*Offer{}
	for _, offer := range aggOffers {
		if offer.Bookmaker == BOOKMAKER && offer.Price < 111 && len(offer.CompetitorPrices) > 4 {
			filteredOffers = append(filteredOffers, offer)
		}
	}

	// limit to 10
	if len(filteredOffers) > 10 {
		filteredOffers = filteredOffers[:10]
	}

	return filteredOffers, nil
}

func getOffersForSport(sportKey string, markets string) ([]*Offer, error) {
	// fetch events from odds api. we'll iterate through these and get more bets
	events, err := getEventsFromOddsApi(sportKey)
	if err != nil {
		return nil, err
	}

	// iterate through events, fetch details, and get best offers from event
	bestOffers := []*Offer{}
	for _, event := range events {
		eventDetails, err := getEventDetailsFromEventOddsApi(event.Id, sportKey, markets)
		if err != nil {
			return nil, err
		}
		bestOffers = append(bestOffers, processEvent(eventDetails)...)
	}
	return bestOffers, nil
}

func getEventsFromOddsApi(sportKey string) ([]Event, error) {
	url := fmt.Sprintf("https://api.the-odds-api.com/v4/sports/%s/odds/?apiKey=%s&regions=us", sportKey, API_KEY)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	var events []Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	fmt.Printf("Successfully fetched %d events from odds API\n", len(events))
	return events, nil
}

func getEventDetailsFromEventOddsApi(eventId string, sportKey string, markets string) (*Event, error) {
	url := fmt.Sprintf("https://api.the-odds-api.com/v4/sports/%s/events/%s/odds?apiKey=%s&regions=us&markets=%s&oddsFormat=american", sportKey, eventId, API_KEY, markets)
	// fmt.Println(url)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	var eventDetails Event
	err = json.Unmarshal(body, &eventDetails)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return &eventDetails, nil
}

func processEvent(event *Event) []*Offer {
	// For this event, find the best odds across all bookmakers
	// We can keep a map:
	//  - Key is a concatenation of market key, outcome name, outcome description, and outcome point
	//  - Value is a list off offers from different bookmakers. Each offer is a bookmaker and a price.
	offerMap := make(map[string][]Offer)
	for _, bookmaker := range event.Bookmakers {
		for _, market := range bookmaker.Markets {
			for _, outcome := range market.Outcomes {
				key := market.Key + outcome.Name + outcome.Description + fmt.Sprintf("%f", outcome.Point)
				offerMap[key] = append(offerMap[key], Offer{
					event.Id,
					event.SportKey,
					event.HomeTeam,
					event.AwayTeam,
					event.CommenceTime,
					bookmaker.Key,
					market.Key,
					outcome.Name,
					outcome.Description,
					outcome.Point,
					int64(outcome.Price),
					convertPriceToDecimal(int64(outcome.Price)),
					0.0,
					0.0,
					nil,
				})
			}
		}
	}

	// Aggregate best offers
	bestOffers := []*Offer{}
	for _, offers := range offerMap {
		var sum float64
		var maxOffer Offer
		for _, offer := range offers {
			sum += offer.DecimalPrice
			if offer.DecimalPrice > maxOffer.DecimalPrice {
				maxOffer = offer
			}
		}
		averageDecimalPrice := sum / float64(len(offers))
		maxOffer.AvergeDecimalPrice = averageDecimalPrice
		maxOffer.OutlierScore = maxOffer.DecimalPrice / averageDecimalPrice
		maxOffer.CompetitorPrices = []*CompetitorPrices{}
		for _, offer := range offers {
			maxOffer.CompetitorPrices = append(maxOffer.CompetitorPrices, &CompetitorPrices{
				offer.Bookmaker,
				offer.Price,
			})
		}
		bestOffers = append(bestOffers, &maxOffer)
	}

	return bestOffers
}

func convertPriceToDecimal(price int64) float64 {
	if price >= 100 {
		return float64(price) / 100
	} else if price <= -100 {
		return 100 / float64(-price)
	}
	return float64(price) / 100
}
