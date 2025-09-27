package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Word struct {
	Id              int64    `db:"id"`
	Spanish         string   `db:"spanish"`
	English         string   `db:"english"` // Stored as comma-separated string
	EnglishPrimary  string   `db:"english_primary"`
	WordType        string   `db:"word_type"`
	English_Translations []string `db:"-"` // Ignore in database; load manually
	Correct         bool     `db:"-"` // Ignore in database
	Peek            bool     `db:"-"` // Ignore in database
}

type User struct {
	Id     int64    `db:"id"`
	Name   string   `db:"name"`
	Emails string   `db:"emails"` // Stored as JSON string
}

type Lesson struct {
	Id      int64  `db:"id"`
	WordIDs string `db:"word_list"` // Stored as comma-separated string of IDs
	Words   []Word `db:"-"`         // Ignore in database; load manually
}

type Result struct {
	WordId  int64 `json:"word_id"`
	Success bool  `json:"success"`
}

type History struct {
	tableName struct{}  `pg:"history"`
	Id        int64     `pg:"id,pk"`
	LessonId  int64     `pg:"lesson_id"`
	Lesson    *Lesson   `pg:"rel:belongs-to"`
	UserId    int64     `pg:"user_id"`
	User      *User     `pg:"rel:belongs-to"`
	Results   []Result  `pg:"results"`
	CreatedAt time.Time `pg:"created_at"`
}


func GetWordByID(db *sql.DB, id int64) (*Word, error) {
	var word Word
	query := `SELECT id, spanish, english, english_primary, word_type FROM words WHERE id = ?`
	
	err := db.QueryRow(query, id).Scan(
		&word.Id,
		&word.Spanish,
		&word.English,
		&word.EnglishPrimary,
		&word.WordType,
	)
	if err != nil {
		return nil, err
	}
	
	return &word, nil
}

func GetLessonByID(db *sql.DB, id int64) (*Lesson, error) {
	var lesson Lesson

	// Get lesson
	err := db.QueryRow("SELECT id, word_list FROM lessons WHERE id = ?", id).Scan(
		&lesson.Id, &lesson.WordIDs,
	)
	if err != nil {
		return nil, err
	}

	// Parse word IDs and query in one go
	wordIDStrs := strings.Split(lesson.WordIDs, ",")
	if len(wordIDStrs) == 0 {
		return &lesson, nil
	}

	// Build query with placeholders
	placeholders := strings.Join(wordIDStrs, ",")
	query := fmt.Sprintf(
		"SELECT id, spanish, english, english_primary, word_type FROM words WHERE id IN (%s)",
		placeholders,
	)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var word Word
		rows.Scan(&word.Id, &word.Spanish, &word.English, &word.EnglishPrimary, &word.WordType)
		lesson.Words = append(lesson.Words, word)
	}

	return &lesson, nil
}


//Returns all lessons from the database, but does not load words, this should be deferred until later as needed
func GetAllLessons(db *sql.DB) ([]Lesson, error) {
	var lessons []Lesson

	query := `SELECT id, word_list FROM lessons`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching lessons: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var lesson Lesson
		err := rows.Scan(&lesson.Id, &lesson.WordIDs)
		if err != nil {
			return nil, fmt.Errorf("error scanning lesson: %w", err)
		}
		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lessons: %w", err)
	}

	return lessons, nil
}


func (u User) String() string {
	return fmt.Sprintf("User<%d %s %v>", u.Id, u.Name, u.Emails)
}

func (s Word) String() string {
	return fmt.Sprintf("Word:\n%d\n%s\n%s\n---\n", s.Id, s.Spanish, s.English)
}
