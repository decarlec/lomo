package main

import (
	"bufio"
	"fmt"
	"learn/spanish/pg_data"
	"os"
	"strings"

	"github.com/go-pg/pg/v10"
)

func ImportWords(db *pg.DB, filePath string) {

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
	var words []pg_data.Word

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
		word := pg_data.Word{
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
		return
	}

	for _, word := range words {
		_, err := db.Model(&word).Insert()
		if err != nil {
			fmt.Printf("Error inserting word (Spanish: %s, English: %s): %v\n", word.Spanish, word.English, err)
			continue
		}
		fmt.Printf("Inserted word: Spanish=%s, English=%s, ID=%d\n", word.Spanish, word.English, word.Id)
	}

	fmt.Printf("Successfully processed %d words.\n", len(words))
}

// CreateLessons creates lessons with 30 unique words each from the vocabulary table
func CreateLessons(db *pg.DB, lessonSize int) (int, error) {
	// Count total words in vocabulary
	count, err := db.Model((*pg_data.Word)(nil)).Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count vocabulary words: %w", err)
	}

	if count < lessonSize {
		return 0, fmt.Errorf("not enough words in vocabulary (%d) to create a lesson of size %d", count, lessonSize)
	}

	// Fetch all word IDs
	var wordIds []int64
	err = db.Model((*pg_data.Word)(nil)).
		Column("id").
		Select(&wordIds)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch word IDs: %w", err)
	}

	// Calculate number of lessons possible
	numLessons := count / lessonSize
	lessonsCreated := 0

	// Create lessons
	for i := 0; i < numLessons; i++ {
		// Select 30 unique word IDs for this lesson
		start := i * lessonSize
		end := start + lessonSize
		lessonWords := wordIds[start:end]

		// Create lesson
		lesson := &pg_data.Lesson{
			WordList: lessonWords,
		}

		// Insert lesson into database
		_, err = db.Model(lesson).Insert()
		if err != nil {
			return lessonsCreated, fmt.Errorf("failed to insert lesson %d: %w", i+1, err)
		}

		lessonsCreated++
		fmt.Printf("Created lesson ID=%d with %d words\n", lesson.Id, len(lesson.WordList))
	}

	if lessonsCreated == 0 {
		return 0, fmt.Errorf("no lessons created; insufficient words")
	}

	return lessonsCreated, nil
}
