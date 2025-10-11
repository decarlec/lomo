package lesson

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"slices"
	"strings"
	"github.com/decarlec/lomo/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LessonModel displays flashcards for a lesson
type LessonModel struct {
	lesson     models.Lesson
	words      []models.Word
	textInput  textinput.Model
	lessonType string
	current    int
}


const SHUFFLE = false

// Special case lesson model for Review lessons
func NewReviewLessonModel(lesson *models.Lesson) (*LessonModel, tea.Cmd) {
	words := lesson.Words
	// Always Shuffle words
	for i := range words {
		j := rand.Intn(i + 1)
		words[i], words[j] = words[j], words[i]
	}

	ti := getLessonInput()
	return &LessonModel{
		lesson:     *lesson,
		words:      words,
		textInput:  ti,
		lessonType: "review",
	}, nil

}

// NewLessonModel creates a LessonModel for a given lesson
func NewLessonModel(lessonId int64) (*LessonModel, tea.Cmd) {
	lesson, err := models.GetLessonByID(lessonId)
	if err != nil {
		log.Fatalf("Error fetching lesson by ID: %v\n", err)
	}

	words := lesson.Words
	if SHUFFLE {
		// Shuffle words
		for i := range words {
			j := rand.Intn(i + 1)
			words[i], words[j] = words[j], words[i]
		}
	}

	ti := getLessonInput()
	return &LessonModel{
		lesson:     *lesson,
		words:      words,
		textInput:  ti,
		lessonType: "normal",
	}, nil
}

// LessonModel methods
func (m LessonModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LessonModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	log.Printf("Updating lessonmodel")

	var currentWord = &m.words[m.current]

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if msg.String() == "/" {
			currentWord.Peek = !currentWord.Peek
			return m, nil
		}
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			return m, tea.Quit
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
				//TODO: go back to lesson menu here
				return nil
			}
		case tea.KeyEnter:
			if m.textInput.Value() == currentWord.EnglishPrimary || checkWord(m.textInput.Value(), currentWord.English_Translations) {
				currentWord.Correct = true
				m.textInput.Placeholder = ""
			} else {
				currentWord.Correct = false
				m.textInput.Placeholder = "try again!"
				m.textInput.PlaceholderStyle.Foreground(lipgloss.Color(input_wrong))
				m.textInput.SetValue("")
			}
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
		for section := range strings.SplitSeq(trimmed, " ") {
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

func getNumCorrect(words []models.Word) int {
	num := 0
	for _, word := range words {
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
	s := ""
	//Title bar
	if m.lessonType == "review" {
		s += fmt.Sprintf("Review - Word %d/%d\n\n", getNumCorrect(m.words), len(m.words))
	} else {
		s += fmt.Sprintf("Lesson %d - Word %d/%d\n\n", m.lesson.Id, getNumCorrect(m.words), len(m.words))
	}

	// Word display
	s += "Spanish: "
	s += lipgloss.NewStyle().Bold(true).UnsetPadding().Foreground(cyan).Render(word.Spanish)
	s += "\n"
	s += m.textInput.View() + "\n"

	// Results
	if word.Correct {
		s += correctStyle("Correct! \n" + translation(word))
	} else if word.Peek {
		s += peekStyle(translation(word))
	}

	//Help text
	s += lipgloss.NewStyle().PaddingTop(1).UnsetBold().Render("\nPress the / key to see answer, left/right to navigate, Ctrl+b to go back, Esc to quit.")
	return style(s)
}

func getLessonInput() textinput.Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 10
	var inputStyle = lipgloss.NewStyle().Foreground(orange).Background(bg)
	ti.PromptStyle = inputStyle
	ti.TextStyle = inputStyle
	ti.Cursor.Style = inputStyle
	ti.PlaceholderStyle = inputStyle
	return ti
}

func translation(word models.Word) string {
	return fmt.Sprintf("Translation: %s \n\nOther translations:\n\t%s", word.EnglishPrimary, strings.Join(word.English_Translations, "\n\t"))
}

func style(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(purple)).
		Background(lipgloss.Color(bg)).
		PaddingTop(2).
		PaddingBottom(2).
		PaddingLeft(4).
		Width(100).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(purple)).
		BorderBackground(lipgloss.Color(bg))
	return style.Render(view)
}

func greenStyle(view string) string {
	var style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(green))

	return style.Render(view)
}

func correctStyle(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(green)).
		Background(bg).
		PaddingTop(2).
		PaddingBottom(2)

	return style.Render(view)
}

func peekStyle(view string) string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(orange)).
		Background(bg).
		PaddingTop(2).
		PaddingBottom(2)

	return style.Render(view)
}

// func headerStyle(view string) string {
// 	var style = lipgloss.NewStyle().Foreground(header_color)
// 	return style.Render(view)
// }
