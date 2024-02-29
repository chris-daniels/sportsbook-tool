package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
****************Bubble Tea UI********************************
************************************************************/
type model struct {
	offers   []*Offer
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
}

func initialModel() model {
	offers, err := FetchOffers()
	if err != nil {
		fmt.Printf("Error fetching offers: %v", err)
		os.Exit(1)
	}

	return model{
		selected: make(map[int]struct{}),
		offers:   offers,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.offers)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Offers\n\n"

	// Iterate over our choices
	for i, choice := range m.offers {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s vs. %s: %s, %s, %s, %.1f, %d \n",
			cursor,
			checked,
			choice.EventHomeTeam,
			choice.EventAwayTeam,
			choice.MarketKey,
			choice.OutcomeName,
			choice.OutcomeDesc,
			choice.OutcomePoint,
			choice.Price)
		s += fmt.Sprintf("\tOutlier Score: %.4f \n", choice.OutlierScore)

		competitorPricesString := ""
		for _, competitorPrice := range choice.CompetitorPrices {
			competitorPricesString += fmt.Sprintf("%s: %d, ", competitorPrice.Bookmaker, competitorPrice.Price)
		}
		s += fmt.Sprintf("\tCompetitorPrices: %s \n", competitorPricesString)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
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

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
