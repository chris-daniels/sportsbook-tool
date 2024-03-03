package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // <------------ here
)

var DB_PW = ""

func InsertOffer(offer *Offer) error {
	// Read DB password from file
	pw, err := os.ReadFile("DBPW.txt")
	if err != nil {
		return err
	}
	DB_PW = string(pw)

	// Connect to the database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		"sportsbook-1.c5wy0ceeqamk.us-east-1.rds.amazonaws.com", 5432, "postgres", pw)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStatement := `
	INSERT INTO bets (id, event_id, sport, home_team, away_team, event_time, bookmaker, market, outcome_name, outcome_desc, price, finalized, won)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, false, false)`
	_, err = db.Exec(sqlStatement, uuid.New().String(), offer.EventId, offer.SportKey, offer.EventHomeTeam, offer.EventAwayTeam, offer.CommenceTime, offer.Bookmaker, offer.MarketKey, offer.OutcomeName, offer.OutcomeDesc, offer.Price)
	if err != nil {
		return err
	}
	return nil
}
