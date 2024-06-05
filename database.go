package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // <------------ here
)

var DB_PW = ""

func InsertOffer(offer *Offer) error {
	// Connect to the database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable dbname=%s",
		"localhost", 5432, "postgres", "postgres", "sb_tool")
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
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println("Offer inserted successfully")
	return nil
}

/*
	CREATE TABLE IF NOT EXISTS bets (
		id UUID PRIMARY KEY,
		event_id VARCHAR(255),
		sport VARCHAR(255),
		home_team VARCHAR(255),
		away_team VARCHAR(255),
		event_time TIMESTAMP,
		bookmaker VARCHAR(255),
		market VARCHAR(255),
		outcome_name VARCHAR(255),
		outcome_desc VARCHAR(255),
		price FLOAT,
		finalized BOOLEAN,
		won BOOLEAN
	)

	INSERT INTO "public"."bets" ("id", "event_id", "sport", "home_team", "away_team", "event_time", "bookmaker", "market", "outcome_name", "outcome_desc", "price", "finalized", "won") VALUES
('7e600ee3-84d2-4d79-98d5-240273ef6dce', 'f49a8ff37ba73f47bbf4d4c64cc1df0e', 'baseball_mlb', 'New York Yankees', 'Minnesota Twins', '2024-06-05 23:05:00', 'fanduel', 'batter_hits', 'Over', 'Giancarlo Stanton', -170, 'f', 'f'),
('3314f087-e1f9-4025-8a18-7160672adc0a', 'f49a8ff37ba73f47bbf4d4c64cc1df0e', 'baseball_mlb', 'New York Yankees', 'Minnesota Twins', '2024-06-05 23:05:00', 'fanduel', 'batter_singles', 'Over', 'Anthony Volpe', -140, 'f', 'f');
*/
