package messages

	import "github.com/decarlec/lomo/models"

// Msg types for transitions
type SwitchToLessonMsg struct {
	LessonId int64
}

type SwitchToReviewMsg struct {
	Lesson models.Lesson
}

type SwitchToLessonMenuMsg struct {}

type SwitchToMenuMsg struct{}
