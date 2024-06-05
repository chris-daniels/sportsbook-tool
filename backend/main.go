package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chris-daniels/sportsbook-tool/database"
	"github.com/chris-daniels/sportsbook-tool/odds_api"
)

/************************************************************
****************Bubble Tea UI********************************
************************************************************/
type model struct {
	offers   []*odds_api.Offer
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
}

func initialModel() model {
	offers, err := odds_api.FetchOffers()
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
				err := database.InsertOffer(m.offers[m.cursor])
				if err != nil {
					fmt.Printf("Error inserting offer: %v", err)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
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
		s += fmt.Sprintf("%s [%s] %s vs. %s at %s:\n\t%s, %s, %s, %.1f, %d \n",
			cursor,
			checked,
			choice.EventHomeTeam,
			choice.EventAwayTeam,
			choice.CommenceTime,
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

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
