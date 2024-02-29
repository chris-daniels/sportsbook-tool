package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
)

const (
	NBA_KEY   = "basketball_nba"
	NCAAB_KEY = "basketball_ncaab"
	NFL_KEY   = "americanfootball_nfl"
	NHL_KEY   = "icehockey_nhl"

	NBA_MARKETS   = "h2h,totals,team_totals,spreads,player_points,player_rebounds,player_assists,player_points_rebounds_assists,player_points_rebounds,player_points_assists,player_rebounds_assists"
	NCAAB_MARKETS = "h2h,totals,team_totals,spreads,player_points,player_rebounds,player_assists,player_points_rebounds_assists,player_points_rebounds,player_points_assists,player_rebounds_assists"
	NFL_MARKETS   = "h2h,totals,team_totals,spreads,player_pass_tds,player_pass_yds,player_pass_completions,player_pass_attempts,player_pass_interceptions,player_pass_longest_completion,player_rush_yds,player_rush_attempts,player_rush_longest,player_receptions,player_reception_yds,player_reception_longest,player_kicking_points,player_field_goals,player_tackles_assists,player_anytime_td"
	NHL_MARKETS   = "h2h,totals,team_totals,spreads,player_points,player_power_play_points,player_assists,player_shots_on_goal,player_total_saves"

	DRAFTKINGS_KEY = "draftkings"
	FANDUEL_KEY    = "fanduel"
)

const (
	MARKETS   = NBA_MARKETS
	SPORT     = NBA_KEY
	BOOKMAKER = FANDUEL_KEY
)

var KEYS_MARKETS = [][]string{
	{NBA_KEY, NBA_MARKETS},
	// {NCAAB_KEY, NCAAB_MARKETS},
	// {NFL_KEY, NFL_MARKETS},
	// {NHL_KEY, NHL_MARKETS},
}

/************************************************************
****************API response structure***********************
************************************************************/
type Event struct {
	Id         string      `json:"id"`
	SportKey   string      `json:"sport_key"`
	HomeTeam   string      `json:"home_team"`
	AwayTeam   string      `json:"away_team"`
	Bookmakers []Bookmaker `json:"bookmakers"`
}

type Bookmaker struct {
	Key     string   `json:"key"`
	Markets []Market `json:"markets"`
}

type Market struct {
	Key      string    `json:"key"`
	Outcomes []Outcome `json:"outcomes"`
}

type Outcome struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Point       float64 `json:"point"`
}

/************************************************************
**************Types for odds computation*********************
************************************************************/
type Offer struct {
	// Pulled off of the event
	EventId       string
	SportKey      string
	EventHomeTeam string
	EventAwayTeam string
	Bookmaker     string
	MarketKey     string
	OutcomeName   string
	OutcomeDesc   string
	OutcomePoint  float64
	Price         int64
	DecimalPrice  float64

	// Computed
	AvergeDecimalPrice float64
	OutlierScore       float64
	CompetitorPrices   []*CompetitorPrices
}

type CompetitorPrices struct {
	Bookmaker string
	Price     int64
}

var API_KEY = ""

func main() {
	// Read API key from file
	apiKey, err := os.ReadFile("TOKEN.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	API_KEY = string(apiKey)

	aggOffers := []*Offer{}
	// iterate over all sports and get offers
	for _, keyMarket := range KEYS_MARKETS {
		offers, err := getOffersForSport(keyMarket[0], keyMarket[1])
		if err != nil {
			fmt.Println("Error:", err)
			return
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

	// print the 10 best offers
	fmt.Println("Best offers:")
	for i := 0; i < 10; i++ {
		// print the offer in a pretty manner
		fmt.Println("***********Offer************")
		fmt.Println("\tEventId:", filteredOffers[i].EventId)
		fmt.Println("\tSportKey:", filteredOffers[i].SportKey)
		fmt.Println("\tEventHomeTeam:", filteredOffers[i].EventHomeTeam)
		fmt.Println("\tEventAwayTeam:", filteredOffers[i].EventAwayTeam)
		fmt.Println("\tBookmaker:", filteredOffers[i].Bookmaker)
		fmt.Println("\tMarketKey:", filteredOffers[i].MarketKey)
		fmt.Println("\tOutcomeName:", filteredOffers[i].OutcomeName)
		fmt.Println("\tOutcomeDesc:", filteredOffers[i].OutcomeDesc)
		fmt.Println("\tOutcomePoint:", filteredOffers[i].OutcomePoint)
		fmt.Println("\tPrice:", filteredOffers[i].Price)
		fmt.Println("\tDecimalPrice:", filteredOffers[i].DecimalPrice)
		fmt.Println("\tAvergeDecimalPrice:", filteredOffers[i].AvergeDecimalPrice)
		fmt.Println("\tOutlierScore:", filteredOffers[i].OutlierScore)
		competitorPricesString := ""
		for _, competitorPrice := range filteredOffers[i].CompetitorPrices {
			competitorPricesString += fmt.Sprintf("%s: %d, ", competitorPrice.Bookmaker, competitorPrice.Price)
		}
		fmt.Println("\tCompetitorPrices:", competitorPricesString)
		fmt.Println("*****************************")

	}
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
	fmt.Println(url)
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
