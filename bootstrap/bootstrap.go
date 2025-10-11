package main

import (
	"bufio"
	"database/sql"
	"encoding/xml"
	"fmt"

	"os"
	"strings"

	"github.com/decarlec/lomo/models"
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
	askForConfirmation("Deleting words.db. Are you sure?")
	err := os.Remove("../words.db")
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to delete existing database: %v\n", err)
		os.Exit(1)
	}

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
		fmt.Printf("Failed to read schema file: %v", err)
	}

	_, err = db.Exec(string(sqlFile))
	if err != nil {
		fmt.Printf("Failed to execute schema: %v", err)
	}

	words := parseLessonWords("1000words.tsv")
	//Processes words from both files
	processWords(db, words)

	//Create all lessons
	createLessons(db, words, 30)

	//Delete words with no lessons
	deleteOrphanWords(db)
}

// Delete words that have no primary translation 
func deleteOrphanWords(db *sql.DB) error {
	_, err := db.Exec("delete from words where english_primary is null or english_primary = ''")
	return err
}

// Inserts database words.
func processWords(db *sql.DB, lessonWords []XmlWord) error {
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
			println("Rolling back transaction due to error:", err)
			tx.Rollback()
		}
	}()

	processedWords := make(map[string]models.Word)
	// Pre-process words for multiple translations
	for _, xLetter := range xDict.Letters {
		for _, xWord := range xLetter.Words {
			mappedWord, exists := processedWords[xWord.Spanish]
			if exists == true {
				//println("Found existing word for", xWord.Spanish, "adding translation", xWord.English)
				mappedWord.English_Translations = append(mappedWord.English_Translations, xWord.English)
				processedWords[xWord.Spanish] = mappedWord
			} else {
				//println("New word for", xWord.Spanish, "with translation", xWord.English)
				newWord := models.Word{}
				newWord.English_Translations = []string{xWord.English}
				newWord.WordType = xWord.Type

				//Add the primary translation if it exists in the tsv file
				for _, word := range lessonWords {
					if word.Spanish == xWord.Spanish {
						newWord.EnglishPrimary = word.English
					}
				}
				newWord.Spanish = xWord.Spanish
				processedWords[xWord.Spanish] = newWord
			}
		}
	}
	fmt.Printf("Processed %d unique words from XML\n", len(processedWords))

	for _, word := range processedWords {
		err := insertWord(tx, word)
		if err != nil {
			fmt.Println("Error inserting word:", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		println("Rolling back transaction due to error:", err)
		return err
	}
	return nil
}

// Try insert a word, takes a word model. If the word already exists, it will just append the english translation
func insertWord(tx *sql.Tx, word models.Word) error {
	// Insert new word
	_, err := tx.Exec(
		"INSERT INTO words (spanish, english_primary, english_translations, word_type) VALUES (?, ?, ?, ?)",
		word.Spanish, word.EnglishPrimary, strings.Join(word.English_Translations, ","), word.WordType)
	return err
}

// CreateLessons creates lessons with 30 unique words each from the vocabulary table
func createLessons(db *sql.DB, words []XmlWord, lessonSize int) error {
	askForConfirmation("This will create lessons in the db, are you sure?")

	// Pre-load all word IDs once instead of querying for each lesson
	wordMap := make(map[string]int64)
	allSpanishWords := make([]string, 0, len(words))

	for _, word := range words {
		allSpanishWords = append(allSpanishWords, word.Spanish)
	}

	placeholders := strings.Repeat("?,", len(allSpanishWords))
	placeholders = strings.TrimRight(placeholders, ",")
	query := fmt.Sprintf("SELECT id, spanish FROM words WHERE spanish IN (%s)", placeholders)

	args := make([]any, len(allSpanishWords))
	for i, w := range allSpanishWords {
		args[i] = w
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to query words: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var spanish string
		if err := rows.Scan(&id, &spanish); err != nil {
			return err
		}
		wordMap[spanish] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}

	fmt.Printf("Map contains %d words\n", len(wordMap))
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	//Go through the word list and create a map of spanish to id
	for i, lessonWords := range chunkWords(words, lessonSize) {
		var wordIDs []int64
		for _, word := range lessonWords {
			if id, exists := wordMap[word.Spanish]; exists {
				wordIDs = append(wordIDs, id)
			}
		}

		//Convert word ids to strings
		wordIDsStr := make([]string, len(wordIDs))
		for i, id := range wordIDs {
			wordIDsStr[i] = fmt.Sprintf("%d", id)
		}
		wordIDsCSV := strings.Join(wordIDsStr, ",")

		//insert lesson into db
		fmt.Printf("Processing lesson with %d words\n", len(lessonWords))
		fmt.Printf("Inserting Ids for words: %s\n", wordIDsCSV)
		res, err := tx.Exec("INSERT INTO lessons (word_ids) VALUES (?)", wordIDsCSV)
		if err != nil {
			return fmt.Errorf("failed to insert lesson %d: %w", i+1, err)
		}
		lessonID, _ := res.LastInsertId()
		fmt.Printf("Created lesson ID=%d with %d words\n", lessonID, len(wordIDs))
	}

	return tx.Commit()
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

	words := make([]XmlWord, 0)

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
