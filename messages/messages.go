package messages

import "learn/spanish/pg_data"

// Msg types for transitions
type SwitchToLessonMsg struct {
	Lesson pg_data.Lesson
}

type SwitchToMenuMsg struct{}
