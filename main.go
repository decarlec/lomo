package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"fmt"
	"learn/spanish/lesson"
	"learn/spanish/messages"
	"log"
	"os"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	tea "github.com/charmbracelet/bubbletea"
)

// AppModel is the parent model managing sub-models
type AppModel struct {
	currentModel tea.Model
	mainMenu     *MainMenuModel
	lessonMenu   *lesson.LessonMenuModel
	lesson       *lesson.LessonModel
}

type MainMenuModel struct {
	choices  []string         // lesson ids
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
}

var db *sql.DB

// initDB initializes the database connection

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "words.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection initialized")
	return nil
}

// AppModel methods
func (m AppModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchToLessonMsg:
		lessonModel, cmd := lesson.NewLessonModel(db, &msg.Lesson)
		m.currentModel = lessonModel
		m.lesson = lessonModel
		return m, cmd
	case messages.SwitchToMenuMsg:
		m.currentModel = m.mainMenu
		return m, nil
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
	if err := initDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

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

func initialModel() MainMenuModel {
	return MainMenuModel{
		// Our to-do list is a grocery list
		choices: []string{"Lessons", "Review"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
	}
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
			db.Close()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			switch m.cursor {
			case 0:
				return lesson.NewLessonMenuModel(db)
			}
		}
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	// The header
	s := "Welcome To Lomo, the language learning cli app! \n\nWhat would you like to do?\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
