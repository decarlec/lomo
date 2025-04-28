package lesson

import (
	"fmt"
	"learn/spanish/messages"
	"learn/spanish/pg_data"
	"log"
	"regexp"
	"slices"
	"strings"

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
	bg                = lipgloss.Color("#000000")
	default_text      = lipgloss.Color("#7168f2")
	header_color      = lipgloss.Color("#928ced")
	input_color       = lipgloss.Color("#db9a3d")
	input_wrong       = lipgloss.Color("#573c17")
	target_word_color = lipgloss.Color("#03fcc6")
	correct_color     = lipgloss.Color("#03fc56")
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
	var inputStyle = lipgloss.NewStyle().Foreground(input_color).Background(bg)
	ti.PromptStyle = inputStyle
	ti.TextStyle = inputStyle
	ti.Cursor.Style = inputStyle
	ti.PlaceholderStyle = inputStyle

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

	var currentWord = &m.words[m.current]

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
				m.textInput.SetValue("")
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
			if checkWord(m.textInput.Value(), currentWord.English) {
				currentWord.Correct = true
				m.textInput.Placeholder = ""
			} else {
				currentWord.Correct = false
				m.textInput.Placeholder = "try again!"
				m.textInput.PlaceholderStyle.Foreground(lipgloss.Color(input_wrong))
				m.textInput.SetValue("")
			}
		case tea.KeyDelete:
			currentWord.Peek = true
			return m, nil
		}

		//handle actual text input
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func checkWord(input string, translations []string) bool {
	if input == "" {
		return false
	}
	//removes all brakets "({[" and their contents
	re := regexp.MustCompile(`[\(\{\[].*?[\}\)\]]`)
	return slices.Contains(translations, input) || slices.ContainsFunc(translations, func(translation string) bool {
		//remove any bracketed descriptive text.
		trimmed := re.ReplaceAllString(translation, "")
		//Split any different words in the section
		for _, section := range strings.Split(trimmed, " ") {
			//remove any leading and trailing punctuation from the words
			noComma := strings.Trim(section, ",;. ")
			if noComma == input {
				log.Printf("Matched on '%s' for input string of '%s' and translations of '%s'", noComma, input, strings.Join(translations, "<<<>>>"))
				//we have a match
				return true
			}
		}
		return false
	})
}

func getNumCorrect(m LessonModel) int {
	num := 0
	for _, word := range m.words {
		if word.Correct {
			num += 1
		}
	}
	return num
}

func (m LessonModel) View() string {
	if len(m.words) == 0 {
		return "No words in this lesson.\nPress b to go back.\n"
	}

	word := m.words[m.current]
	s := fmt.Sprintf("Lesson %d - Word %d/%d\n\n", m.lesson.Id, getNumCorrect(m), len(m.words))
	s += "Spanish: "
	s += lipgloss.NewStyle().Bold(true).UnsetPadding().Foreground(target_word_color).Render(word.Spanish)
	s += "\n"
	s += m.textInput.View() + "\n"
	if word.Correct {
		s += correctStyle("Correct! \n" + translation(word))
	} else if word.Peek {
		s += peekStyle(translation(word))
	}
	s += lipgloss.NewStyle().PaddingTop(1).UnsetBold().Render("\nPress delete to see answer, left/right to navigate, Ctrl+b to go back, Esc to quit.")
	return style(s)
}

func translation(word pg_data.Word) string {
	return fmt.Sprintf("Translations:\n\t%s", strings.Join(word.English, "\n\t"))
}

func style(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(default_text)).
		Background(lipgloss.Color(bg)).
		PaddingTop(2).
		PaddingBottom(2).
		PaddingLeft(4).
		Width(100).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(default_text)).
		BorderBackground(lipgloss.Color(bg))
	return style.Render(view)
}

func correctStyle(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(correct_color)).
		Background(bg).
		PaddingTop(2).
		PaddingBottom(2)

	return style.Render(view)
}

func peekStyle(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(input_color)).
		Background(bg).
		PaddingTop(2).
		PaddingBottom(2)

	return style.Render(view)
}

func headerStyle(view string) string {
	var style = lipgloss.NewStyle().Foreground(header_color)
	return style.Render(view)
}
