package lesson

import (
	"fmt"
	"learn/spanish/messages"
	"learn/spanish/pg_data"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-pg/pg/v10"
)

// LessonModel displays flashcards for a lesson
type LessonModel struct {
	lesson    pg_data.Lesson
	words     []pg_data.Word
	textInput textinput.Model
	current   int
}

var (
	bg = lipgloss.Color("#7168f2")
	fg = lipgloss.Color("#db9a3d")
)

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

	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 10
	var inputStyle = lipgloss.NewStyle().Foreground(fg).Background(bg)
	ti.PromptStyle = inputStyle
	ti.TextStyle = inputStyle
	ti.Cursor.Style = inputStyle

	return &LessonModel{
		lesson:    lesson,
		words:     words,
		textInput: ti,
	}, nil
}

// LessonModel methods
func (m LessonModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LessonModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			return m, func() tea.Msg {
				return messages.SwitchToMenuMsg{}
			}
		//Scroll words
		case tea.KeyRight:
			if m.current < len(m.words)-1 {
				m.current++
				return m, nil
			}
		//Scroll words
		case tea.KeyLeft:
			if m.current > 0 {
				m.current--
				return m, nil
			}
		//Go back
		case tea.KeyCtrlB:
			return m, func() tea.Msg {
				return messages.SwitchToMenuMsg{}
			}
		case tea.KeyEnter:
			if m.textInput.Value() == m.words[m.current].English {
				m.words[m.current].Correct = true
				m.textInput.SetValue("")
				m.textInput.Placeholder = ""
				m.current++
			} else {
				m.textInput.SetValue("")
				m.textInput.Placeholder = "try again!"
			}
		case tea.KeyDelete:
			m.textInput.SetValue(m.words[m.current].English)
		}

		//handle actual text input
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m LessonModel) View() string {
	if len(m.words) == 0 {
		return "No words in this lesson.\nPress b to go back.\n"
	}

	word := m.words[m.current]
	s := fmt.Sprintf("Lesson %d - Word %d/%d\n\n", m.lesson.Id, m.current+1, len(m.words))
	s += fmt.Sprintf("Spanish: %s\n", word.Spanish)
	s += m.textInput.View() + "\n"
	if word.Correct {
		s += fmt.Sprintf("Correct! %s means %s!\n", word.Spanish, word.English)
	}
	s += "\nPress space to flip, left/right to navigate, Ctrl+b to go back, Esc to quit.\n"
	return style(s)
}

func style(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(bg).
		PaddingTop(2).
		PaddingLeft(4).
		Width(100)

	return style.Render(view)
}
