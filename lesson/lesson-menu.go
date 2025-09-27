package lesson

import (
	"fmt"
	"log"
	"learn/spanish/messages"
	"learn/spanish/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"database/sql"
)

// MenuModel displays a list of lessons
type LessonMenuModel struct {
	lessons []models.Lesson
	review  models.Lesson
	cursor  int
}

// NewMenuModel creates a MenuModel with lessons from the database
func NewLessonMenuModel(db *sql.DB) (*LessonMenuModel, tea.Cmd) {
	lessons, err := models.GetAllLessons(db)

	if err != nil {
		log.Fatalf("Error fetching lessons: %v\n", err)
	}

	return &LessonMenuModel{lessons: lessons}, nil
}

// MenuModel methods
func (m LessonMenuModel) Init() tea.Cmd {
	return nil
}

func (m LessonMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.lessons)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.lessons) {
				return m, func() tea.Msg {
					return messages.SwitchToLessonMsg{Lesson: m.lessons[m.cursor]}
				}
			}
		}
	}
	return m, nil
}

func (m LessonMenuModel) View() string {
	s := "Select a Lesson:\n\n"
	for i, lesson := range m.lessons {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		var lessonStr string
		if i == 0 {
			lessonStr = fmt.Sprintf("%s Review\n", cursor)
		} else {
			lessonStr = fmt.Sprintf("%s Lesson %d (%d/%d) \n", cursor, lesson.Id, getNumCorrect(lesson.Words), len(lesson.WordIDs))
		}
		if getNumCorrect(lesson.Words) == len(lesson.WordIDs) {
			s += greenStyle(lessonStr)
		} else {
			s += lipgloss.NewStyle().Render(lessonStr)
		}
	}
	s += "\nPress q to quit.\n"
	return s
}
