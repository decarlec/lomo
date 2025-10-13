package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/decarlec/lomo/db"
)

type Word struct {
	Id              int64    `db:"id"`
	Spanish         string   `db:"spanish"`
	EnglishTranslations       string   `db:"english_translations"` // Stored as comma-separated string
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
	WordIDs string `db:"word_ids"` // Stored as comma-separated string of IDs
	Words   []Word `db:"-"`         // Ignore in database; load manually
}

type Result struct {
	WordId  int64 `json:"word_id"`
	Success bool  `json:"success"`
}

type History struct {
	Id        int64     `db:"id"`
	LessonId  int64     `db:"lesson_id"`
	UserId    int64     `db:"user_id"`
	CorrectIds string `db:"correct_ids"`
	CreatedAt time.Time `db:"created_at"`
}


func GetWordByID(db *sql.DB, id int64) (Word, error) {
	var word Word
	query := `SELECT id, spanish, english_translations, english_primary, word_type FROM words WHERE id = ?`
	
	err := db.QueryRow(query, id).Scan(
		&word.Id,
		&word.Spanish,
		&word.EnglishTranslations,
		&word.EnglishPrimary,
		&word.WordType,
	)
	if err != nil {
		return Word{}, err
	}
	
	return word, nil
}

func GetAllWords() ([]Word, error) {
	db := db.DB;
	query := "SELECT id, spanish, english_translations, english_primary, word_type FROM words"

	words := []Word{}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var word Word
		rows.Scan(&word.Id, &word.Spanish, &word.EnglishTranslations, &word.EnglishPrimary, &word.WordType)
		//Need to process the English translations into a slice
		word.English_Translations = strings.Split(word.EnglishTranslations, ",")
		words = append(words, word)
	}

	return words, nil
}

func GetLessonByID(id int64) (*Lesson, error) {
	var lesson Lesson

	// Get lesson
	err := db.DB.QueryRow("SELECT id, word_ids FROM lessons WHERE id = ?", id).Scan(
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
		"SELECT id, spanish, english_translations, english_primary, word_type FROM words WHERE id IN (%s)",
		placeholders,
	)

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var word Word
		rows.Scan(&word.Id, &word.Spanish, &word.EnglishTranslations, &word.EnglishPrimary, &word.WordType)
		//Need to process the English translations into a slice
		word.English_Translations = strings.Split(word.EnglishTranslations, ",")
		lesson.Words = append(lesson.Words, word)
	}

	return &lesson, nil
}


//Returns all lessons from the database, but does not load words, this should be deferred until later as needed
func GetAllLessons() ([]Lesson, error) {
	var lessons []Lesson

	query := `SELECT id, word_ids FROM lessons`
	
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching lessons: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		log.Println("Scanning lesson row")
		var lesson Lesson
		err := rows.Scan(&lesson.Id, &lesson.WordIDs)
		if err != nil {
			return nil, fmt.Errorf("error scanning lesson: %w", err)
		}
		log.Println("Appending lesson")
		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lessons: %w", err)
	}

	return lessons, nil
}

func GetHistoryForLesson(lessonId int64) ([]History, error) {
	histories := []History{}
	query := `SELECT id, lesson_id, user_id, correct_ids, created_at FROM history WHERE lesson_id = ?`
	// Get lesson

	rows, err := db.DB.Query(query, lessonId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var history History
		rows.Scan(&history.Id, &history.LessonId, &history.CorrectIds, &history.CreatedAt, &history.UserId)
		//Need to process the English translations into a slice

		histories = append(histories, history)
	}

	return histories, nil
}

func WriteHistory(lesson Lesson, userId int, words []Word) error {
	correctIds := ""
	for wordId := range lesson.Words {
		if words[wordId].Correct {
		correctIds += fmt.Sprintf("%d,", lesson.Words[wordId].Id)
	}
	}
	log.Printf("Writing history for lesson %d, user %d\n, words %s ", lesson.Id, userId, correctIds)
	_, err := db.DB.Exec("INSERT INTO history (lesson_id, user_id, correct_ids) VALUES (?, ?, ?)", lesson.Id, userId, correctIds)
	return  err 
}


func (u User) String() string {
	return fmt.Sprintf("User<%d %s %v>", u.Id, u.Name, u.Emails)
}

func (s Word) String() string {
	return fmt.Sprintf("Word:\n%d\n%s\n%s\n---\n", s.Id, s.Spanish, s.EnglishPrimary)
}
