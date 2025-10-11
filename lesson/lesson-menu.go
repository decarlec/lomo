package lesson

import (
	"fmt"
	"log"
	"strings"

	"github.com/decarlec/lomo/messages"
	"github.com/decarlec/lomo/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// MenuModel displays a list of lessons
type LessonMenuModel struct {
	lessons []models.Lesson
	cursor  int
}

var (

    headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(purple)).Bold(true).Align(lipgloss.Center)
    cellStyle    = lipgloss.NewStyle().Padding(0, 1).Width(10).Bold(true).AlignHorizontal(lipgloss.Center)
    oddRowStyle  = cellStyle.Foreground(orange)
    evenRowStyle = cellStyle.Foreground(orange_wash)
)

// NewMenuModel creates a MenuModel with lessons from the database
func NewLessonMenuModel() (*LessonMenuModel, tea.Cmd) {
	lessons, err := models.GetAllLessons()

	if err != nil {
		log.Fatalf("Error fetching lessons: %v\n", err)
	}	
	log.Printf("Fetched %d lessons from database\n", len(lessons))
	return &LessonMenuModel{lessons: lessons }, nil
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
		case "esc":
			return m, func() tea.Msg {
				log.Printf("Switching to lesson %d\n", m.lessons[m.cursor].Id)
					return messages.SwitchToMenuMsg{}
				}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.lessons)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				log.Printf("Switching to lesson %d\n", m.lessons[m.cursor].Id)
					return messages.SwitchToLessonMsg{LessonId: m.lessons[m.cursor].Id}
				}
		}
	}
	return m, nil
}

func (m LessonMenuModel) View() string {
	rows := make([][]string, len(m.lessons))
	for lessonIndex, lesson := range m.lessons {
		cursor := "  "
		if m.cursor == lessonIndex {
			cursor = "=>"
		}
		rows[lessonIndex] = []string{fmt.Sprintf("%s %d", cursor, lesson.Id), fmt.Sprintf("%d/%d", getNumCorrect(lesson.Words), len(strings.Split(lesson.WordIDs, ",")))}
	}
	table := table.New().
    Border(lipgloss.RoundedBorder()).
    BorderStyle(
			lipgloss.NewStyle().Foreground(purple).
			Bold(true)).
    StyleFunc(func(row, col int) lipgloss.Style {
        switch {
        case row == table.HeaderRow:
            return headerStyle
        case row%2 == 0:
            return evenRowStyle
        default:
            return oddRowStyle
        }
    }).
    Headers("Lesson", "Progress").
    Rows(rows...).Render()

		return table + "\nPress q to quit.\n"
}
