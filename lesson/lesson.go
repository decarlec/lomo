package lesson

import (
	"fmt"
	"learn/spanish/messages"
	"learn/spanish/pg_data"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-pg/pg/v10"
)

// LessonModel displays flashcards for a lesson
type LessonModel struct {
	lesson    pg_data.Lesson
	words     []pg_data.Word
	current   int
	showFront bool
}

// NewLessonModel creates a LessonModel for a given lesson
func NewLessonModel(db *pg.DB, lesson pg_data.Lesson) (*LessonModel, tea.Cmd) {
	if db == nil {
		return &LessonModel{}, tea.Quit
	}

	var words []pg_data.Word
	err := db.Model(&words).Where("id IN (?)", pg.In(lesson.WordList)).Select()
	if err != nil {
		fmt.Printf("Error fetching words: %v\n", err)
		return &LessonModel{}, tea.Quit
	}

	return &LessonModel{
		lesson:    lesson,
		words:     words,
		showFront: true,
	}, nil
}

// LessonModel methods
func (m LessonModel) Init() tea.Cmd {
	return nil
}

func (m LessonModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, func() tea.Msg {
				return messages.SwitchToMenuMsg{}
			}
		case "right", "l":
			if m.current < len(m.words)-1 {
				m.current++
				m.showFront = true
			}
		case "left", "h":
			if m.current > 0 {
				m.current--
				m.showFront = true
			}
		case " ":
			m.showFront = !m.showFront
		case "b":
			return m, func() tea.Msg {
				return messages.SwitchToMenuMsg{}
			}
		}
	}
	return m, nil
}

func (m LessonModel) View() string {
	if len(m.words) == 0 {
		return "No words in this lesson.\nPress b to go back.\n"
	}

	word := m.words[m.current]
	s := fmt.Sprintf("Lesson %d - Word %d/%d\n\n", m.lesson.Id, m.current+1, len(m.words))
	if m.showFront {
		s += fmt.Sprintf("Spanish: %s\n", word.Spanish)
	} else {
		s += fmt.Sprintf("English: %s\n", word.English)
	}
	s += "\nPress space to flip, h/l to navigate, b to go back, q to quit.\n"
	return s
}
