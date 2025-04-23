package main

import (
	"fmt"
	"learn/spanish/pg_data"

	"github.com/go-pg/pg/v10"
)

func test() {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "example",
		Database: "postgres",
		Addr:     "localhost:5432", // Adjust if Docker exposes a different port
	})
	defer db.Close()

	// Insert a word
	word := &pg_data.Word{Spanish: "hola", English: "hello"}
	_, err := db.Model(word).Insert()
	if err != nil {
		panic(err)
	}

	fmt.Println(word.String())

	// Insert a lesson
	lesson := &pg_data.Lesson{WordList: []int64{1, word.Id}}
	_, err = db.Model(lesson).Insert()
	if err != nil {
		panic(err)
	}

	// Load lesson with words
	loadedLesson := &pg_data.Lesson{}
	err = db.Model(loadedLesson).
		Where("lesson.id = ?", lesson.Id).
		Select()
	if err != nil {
		panic(err)
	}

	// Load related words using word_list
	if len(loadedLesson.WordList) > 0 {
		err = db.Model(&loadedLesson.Words).
			Where("word.id IN (?)", pg.In(loadedLesson.WordList)).
			Select()
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Loaded lesson: %+v\nWords: %+v\n", loadedLesson, loadedLesson.Words)
}
