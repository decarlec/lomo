package main

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"learn/spanish/pg_data"
	"os"
	"strings"

	"github.com/go-pg/pg/v10"
)

func main() {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "example",
		Database: "postgres",
		Addr:     "localhost:5432", // Adjust if Docker exposes a different port
	})

	_, err := db.Exec("SELECT 1")
	if err != nil {
		panic(err)
	}

	ProcessWords(db)
	CreateLessons(db, parseLessonWords("1000words.tsv"), 30)
}

// Parse words from file (1000 words.txt) and return array of words
func parseLessonWords(filePath string) []Word {
	// Open the TSV file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read the file
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var words []Word

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Split the line by tabs
		fields := strings.Split(line, "\t")
		if lineNumber == 1 {
			// Verify headers
			if len(fields) != 3 || fields[0] != "Number" || fields[1] != "Spanish" || fields[2] != "in English" {
				fmt.Println("Error: Invalid header format. Expected 'Number\tSpanish\tin English'")
				os.Exit(1)
			}
			continue // Skip header row
		}

		// Validate row
		if len(fields) != 3 {
			fmt.Printf("Error: Invalid row format at line %d: %s\n", lineNumber, line)
			continue
		}

		// Create Word struct
		word := Word{
			Spanish: strings.TrimSpace(fields[1]),
			English: strings.TrimSpace(fields[2]),
		}

		// Skip empty Spanish or English fields
		if word.Spanish == "" || word.English == "" {
			fmt.Printf("Warning: Skipping row at line %d due to empty Spanish or English: %s\n", lineNumber, line)
			continue
		}

		words = append(words, word)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Insert words into the database
	if len(words) == 0 {
		fmt.Println("No valid words to insert.")
		return nil
	}

	fmt.Println("Finished processing words.")
	return words
}

// chunkWords splits a slice of words into chunks of lessonSize
func chunkWords(words []Word, lessonSize int) [][]Word {
	var chunks [][]Word
	for i := 0; i < len(words); i += lessonSize {
		end := i + lessonSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, words[i:end])
	}
	return chunks
}

// CreateLessons creates lessons with 30 unique words each from the vocabulary table
func CreateLessons(db *pg.DB, words []Word, lessonSize int) error {
	askForConfirmation("This will create lessons in the db, are you sure?")
	// Create lessons
	for i, lessonWords := range chunkWords(words, lessonSize) {
		//Get the spanish words for the query
		var spanishWords []string
		for _, word := range lessonWords {
			spanishWords = append(spanishWords, word.Spanish)
		}

		//Get the words from the db
		var dbWords []*pg_data.Word
		err := db.Model((&dbWords)).
			Where("spanish IN (?)", pg.In(spanishWords)).
			Select()

		if len(dbWords) > 0 {
			fmt.Println(spanishWords)
			fmt.Println(dbWords)
		}

		if err != nil {
			panic(err)
		}

		//Convert words to a word id list
		wordList := []int64{}
		for _, dbWord := range dbWords {
			fmt.Println(dbWord.Spanish)
			wordList = append(wordList, dbWord.Id)
		}

		// Create lesson with word ids
		lesson := &pg_data.Lesson{
			WordList: wordList,
		}

		// Insert lesson into database
		_, err = db.Model(lesson).Insert()
		if err != nil {
			return fmt.Errorf("failed to insert lesson %d: %w", i+1, err)
		}

		fmt.Printf("Created lesson ID=%d with %d words\n", lesson.Id, len(lesson.WordList))
	}

	return nil
}

type Dictionary struct {
	Letters []Letter `xml:"l"`
}

type Letter struct {
	Words []Word `xml:"w"`
}

type Word struct {
	Spanish string `xml:"c"`
	English string `xml:"d"`
	Type    string `xml:"t"`
}

// UpsertWord checks if a Spanish word exists, creates it if not, or appends English translation if it does
func upsertWord(db *pg.DB, spanishWord, englishTranslation, wordType string) error {
	// Use a transaction to ensure atomicity
	return db.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		// Attempt to insert or update using ON CONFLICT
		query := `
			INSERT INTO words (spanish, english, word_type)
			VALUES (?, ARRAY[?]::text[], ?)
			ON CONFLICT (spanish) DO UPDATE
			SET english = array_append(words.english, ?)
			RETURNING spanish
		`
		var resultSpanish pg_data.Word
		_, err := tx.QueryOne(&resultSpanish, query, spanishWord, englishTranslation, wordType, englishTranslation)
		if err != nil {
			panic(err)
		}
		return nil
	})
}

// Inserts database words.
func ProcessWords(db *pg.DB) error {
	dat, err := os.ReadFile("es-en.xml")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	askForConfirmation("This will put a bunch of words in the db, and take a while. Are you sure?")

	var dict Dictionary

	if err := xml.Unmarshal([]byte(dat), &dict); err != nil {
		panic(err)
	}
	for _, letter := range dict.Letters {
		fmt.Println(letter.Words[0])
		for _, word := range letter.Words {
			// Assuming word.English is a slice, take first
			err := upsertWord(db, word.Spanish, word.English, word.Type)
			if err != nil {
				fmt.Println("UPSERT ERROR")
			}
		}
	}
	return nil
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", s)
	response, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return true
	}
	return false
}
