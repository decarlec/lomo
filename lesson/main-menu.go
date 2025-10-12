package lesson

import (
	"fmt"
	"log"

	"github.com/decarlec/lomo/assets"
	"github.com/decarlec/lomo/db"
	"github.com/decarlec/lomo/messages"
	"github.com/decarlec/lomo/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

type MainMenuModel struct {
	Choices  []string         // lesson ids
	cursor   int              // which to-do list item our cursor is pointing at
	Selected map[int]struct{} // which to-do items are selected
	Logo string	
}

func (m MainMenuModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			db.DB.Close()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.Selected[m.cursor]
			if ok {
				delete(m.Selected, m.cursor)
			} else {
				m.Selected[m.cursor] = struct{}{}
			}
			switch m.cursor {
			case 0:
				return m, func() tea.Msg {
					log.Printf("Switching to lesson menu\n")
					return messages.SwitchToLessonMenuMsg{}
				}
			case 1:
				return m, func() tea.Msg {
					log.Printf("Switching to Review Lesson\n")
					return messages.SwitchToReviewMsg{Lesson: *getReviewLesson()}
				}
			}

		}
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	// The header
	var welcome = lipgloss.NewStyle().Foreground(lipgloss.Color(assets.Orange)).Bold(true).Align(lipgloss.Center).Render(
`Welcome to
`)

	var message = lipgloss.NewStyle().Foreground(lipgloss.Color(assets.Orange)).Bold(true).Align(lipgloss.Center).Render(
`

	...the language learning TUI (too-ee) app!
`)

	s := welcome + m.Logo + message + "\n"

	// Iterate over our choices
	for i, choice := range m.Choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "=>"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

		// Render the row

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(assets.Cyan)).
		PaddingTop(2).
		PaddingLeft(4).
		Width(100).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(assets.Purple))


	s = style.Render(s)

	var exit = "\nPress q to quit.\n"

	s += exit

	// Send the UI for rendering
	return s
}

func getReviewLesson() *models.Lesson {
	words, err := models.GetAllWords()
	if err != nil {
		log.Fatalf("Error fetching all words for review lesson: %v\n", err)
	}

	return &models.Lesson{Words: words}
}
