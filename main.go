package main

import (
	"fmt"
	"os"

	"github.com/decarlec/lomo/db"
	"github.com/decarlec/lomo/lesson"
	"github.com/decarlec/lomo/messages"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
)

// AppModel is the parent model managing sub-models
type AppModel struct {
	currentModel tea.Model
	mainMenu     *lesson.MainMenuModel
	lessonMenu   *lesson.LessonMenuModel
	lesson       *lesson.LessonModel
	lessonsInProgress []lesson.LessonModel
	review *lesson.LessonModel
}

// AppModel methods
func (m AppModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchToLessonMsg:
		//Select in progress lesson if possible
		for _, l := range m.lessonsInProgress {
			if l.Lesson.Id == msg.LessonId {
				m.lessonsInProgress = append(m.lessonsInProgress, l)
				m.lesson = &l
				m.currentModel = m.lesson
				return m, nil
			}
		}
		lessonModel, cmd := lesson.NewLessonModel(msg.LessonId)
		m.currentModel = lessonModel
		m.lesson = lessonModel
		return m, cmd
	case messages.SwitchToMenuMsg:
		m.currentModel = m.mainMenu
		return m, nil
	case messages.SwitchToLessonMenuMsg:
		if m.lessonMenu == nil {
			m.lessonMenu, _ = lesson.NewLessonMenuModel()
		}
		m.currentModel = m.lessonMenu
	case messages.SwitchToReviewMsg:
		if m.review == nil {
			m.review, _ = lesson.NewReviewLessonModel(&msg.Lesson)
		}
		m.currentModel = m.review
	}
	var cmd tea.Cmd
	m.currentModel, cmd = m.currentModel.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	return m.currentModel.View()
}

func main() {
	// Setup logging
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Connect to db
	if err := db.InitDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	// Initialize models
	mainMenu := initialModel()
	appModel := AppModel{
		currentModel: mainMenu,
		mainMenu:     &mainMenu,
	}
	p := tea.NewProgram(appModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() lesson.MainMenuModel {
	return lesson.MainMenuModel{
		// Our to-do list is a grocery list
		Choices: []string{"Lessons", "Review"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		Selected: make(map[int]struct{}),
	}
}

