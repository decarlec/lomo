package messages

import "learn/spanish/models"

// Msg types for transitions
type SwitchToLessonMsg struct {
	Lesson models.Lesson
}

type SwitchToReviewMsg struct {
	Lesson models.Lesson
}

type SwitchToMenuMsg struct{}
