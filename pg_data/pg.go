package pg_data

import (
	"fmt"
	"time"
)

type Word struct {
	tableName struct{} `pg:"words"`
	Id        int64    `pg:"id,pk"`
	Spanish   string   `pg:"spanish"`
	English   string   `pg:"english"`
}

type User struct {
	tableName struct{} `pg:"users"`
	Id        int64    `pg:"id,pk"`
	Name      string   `pg:"name"`
	Emails    []string `pg:"emails"`
}

type Lesson struct {
	tableName struct{} `pg:"lesson"`
	Id        int64    `pg:"id,pk"`
	WordList  []int64  `pg:"word_list,array"`
	Words     []Word   `pg:"-"` // Ignore in SQL; load manually
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

func (u User) String() string {
	return fmt.Sprintf("User<%d %s %v>", u.Id, u.Name, u.Emails)
}

func (s Word) String() string {
	return fmt.Sprintf("Word<%d %s %s>", s.Id, s.Spanish, s.English)
}
