package messages

// Msg types for transitions
type SwitchToLessonMsg struct {
	LessonId int64
}

type SwitchToReviewMsg struct {}

type SwitchToLessonMenuMsg struct {}

type SwitchToMenuMsg struct{}
