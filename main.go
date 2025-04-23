package main

// These imports will be used later on the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.
import (
	"fmt"
	"learn/spanish/lesson"
	"learn/spanish/messages"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-pg/pg/v10"
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

var db *pg.DB

// initDB initializes the database connection

func initDB() error {
	db = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "example",
		Database: "postgres",
		Addr:     "localhost:5432", // Adjust if Docker exposes a different port
	})

	_, err := db.Exec("SELECT 1")
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Database connection initialized")
	return nil
}

// AppModel methods
func (m AppModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SwitchToLessonMsg:
		lessonModel, cmd := lesson.NewLessonModel(db, msg.Lesson)
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
	if err := initDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

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
		choices: []string{"Test db", "Bootstrap", "Lessons"},

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

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			db.Close()
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

			switch m.cursor {
			case 0:
				test()
			case 1:
				ImportWords(db, "1000words.tsv")
				CreateLessons(db, 30)
			case 2:
				return lesson.NewLessonMenuModel(db)
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m MainMenuModel) View() string {
	// The header
	s := "Welcome! \n\nWhat would you like to do?\n"

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
