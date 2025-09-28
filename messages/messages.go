package messages

import "learn/spanish/models"

// Msg types for transitions
type SwitchToLessonMsg struct {
	LessonId int64
}

type SwitchToReviewMsg struct {
	Lesson models.Lesson
}

type SwitchToMenuMsg struct{}
