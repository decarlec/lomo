package main

import (
	"bufio"
	"database/sql"
	"encoding/xml"
	"fmt"
	"learn/spanish/models"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type XmlDictionary struct {
	Letters []XmlLetter `xml:"l"`
}

type XmlLetter struct {
	Words []XmlWord `xml:"w"`
}

type XmlWord struct {
	Spanish string `xml:"c"`
	English string `xml:"d"`
	Type    string `xml:"t"`
}

func main() {
	// Initialize SQLite3 database
	db, err := sql.Open("sqlite3", "../words.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Init databse schema
	sqlFile, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}

	_, err = db.Exec(string(sqlFile))
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	words := parseLessonWords("1000words.tsv")
	processWords(db)
	//First lesson will be a review of all words.
	createLessons(db, words, 1000)
	createLessons(db, words, 30)
}
//
// Inserts database words.
func processWords(db *sql.DB) error {
	dat, err := os.ReadFile("es-en.xml")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	askForConfirmation("This will put a bunch of words in the db, and take a while. Are you sure?")

	var xDict XmlDictionary

	if err := xml.Unmarshal([]byte(dat), &xDict); err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	processedWords := make(map[string]models.Word)
	// Pre-process words for multiple translations
	for _, xLetter := range xDict.Letters {
		for _, xWord := range xLetter.Words {
			mappedWord, exists := processedWords[xWord.Spanish]
			if exists == true {
				mappedWord.English_Translations = append(mappedWord.English_Translations, xWord.English)
			} else {
				mappedWord.EnglishPrimary = xWord.English
				mappedWord.English_Translations = []string{xWord.English}
				mappedWord.WordType = xWord.Type
			}
			mappedWord.Spanish = xWord.Spanish

			processedWords[xWord.Spanish] = mappedWord
		}
	}

	for _, word := range processedWords {
		if word.Spanish == "hecho" {
			log.Println(word.Spanish, word.EnglishTranslations)
		}
			err := insertWord(tx, word)
			if err != nil {
				fmt.Println("UPSERT ERROR:", err)
				return err
			}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Try insert a word, takes a word model. If the word already exists, it will just append the english translation 
func insertWord(tx *sql.Tx, word models.Word) error {
    var existingEnglish string
    err := tx.QueryRow("SELECT english_translations FROM words WHERE spanish = ?", word.Spanish).Scan(&existingEnglish)
    
    if err == sql.ErrNoRows {
        // Insert new word
        _, err = tx.Exec(
					"INSERT INTO words (spanish, english_primary, english_translations, word_type, ) VALUES (?, ?, ?, ?)",
            word.Spanish, word.EnglishPrimary, strings.Join(word.English_Translations, ","), word.WordType)
        return err
    } else if err != nil {
        return err
    }
    
    // Spanish word exists - always append the new translation
    newEnglish := existingEnglish + "," + word.EnglishTranslations

    _, err = tx.Exec(
        "UPDATE words SET english_translations = ? WHERE spanish = ?",
        newEnglish, word.Spanish,
    )
    return err
}

//Updated the primary translation for a word, inserting if it doesn't exist
func updatePrimaryTranslation(db *sql.DB, spanishWord, translation string) error {
	// Use a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Try to update first
	res, err := tx.Exec("UPDATE words SET english = ? WHERE spanish = ?", translation, spanishWord)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Insert if not exists
		_, err = tx.Exec("INSERT INTO words (spanish, english) VALUES (?, ?)", spanishWord, translation)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}


// CreateLessons creates lessons with 30 unique words each from the vocabulary table
func createLessons(db *sql.DB, words []XmlWord, lessonSize int) error {
	// First update the primary translations in the db
	for _, word := range words {
		err := updatePrimaryTranslation(db, word.Spanish, word.English)
		if err != nil {
			fmt.Printf("Failed to update translation for %s: %v\n", word.Spanish, err)
		}
	}

	askForConfirmation("This will create lessons in the db, are you sure?")

	// Begin transaction for all lesson inserts
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for i, lessonWords := range chunkWords(words, lessonSize) {
		// ...existing code to build spanishWords, query word IDs, etc...

		// Prepare query to get word IDs
		var spanishWords []string
		for _, word := range lessonWords {
			spanishWords = append(spanishWords, word.Spanish)
		}
		placeholders := strings.Repeat("?,", len(spanishWords))
		placeholders = strings.TrimRight(placeholders, ",")
		query := fmt.Sprintf("SELECT id, spanish FROM words WHERE spanish IN (%s)", placeholders)

		args := make([]any, len(spanishWords))
		for i, w := range spanishWords {
			args[i] = w
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			return fmt.Errorf("failed to query words: %w", err)
		}
		defer rows.Close()

		var wordIDs []int64
		for rows.Next() {
			var id int64
			var spanish string
			if err := rows.Scan(&id, &spanish); err != nil {
				return err
			}
			wordIDs = append(wordIDs, id)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		wordIDsStr := make([]string, len(wordIDs))
		for i, id := range wordIDs {
			wordIDsStr[i] = fmt.Sprintf("%d", id)
		}
		wordIDsCSV := strings.Join(wordIDsStr, ",")

		// Insert lesson using transaction
		res, err := tx.Exec("INSERT INTO lessons (word_ids) VALUES (?)", wordIDsCSV)
		if err != nil {
			return fmt.Errorf("failed to insert lesson %d: %w", i+1, err)
		}
		lessonID, _ := res.LastInsertId()
		fmt.Printf("Created lesson ID=%d with %d words\n", lessonID, len(wordIDs))
	}

	// Commit transaction after all lessons are inserted
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Parse words from file (1000 words.txt) and return array of words
func parseLessonWords(filePath string) []XmlWord {
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
	var words []XmlWord

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
		word := XmlWord{
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

// chunkWords splits a slice of words into chunks of lessonSize
func chunkWords(words []XmlWord, lessonSize int) [][]XmlWord {
	var chunks [][]XmlWord
	for i := 0; i < len(words); i += lessonSize {
		end := min(i+lessonSize, len(words))
		chunks = append(chunks, words[i:end])
	}
	return chunks
}
