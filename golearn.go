package golearn

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// ModeTyping constant for user "typing" mode
const ModeTyping = "typing"

// ModePicking constant for user "picking" mode
const ModePicking = "picking"

// LogFormatter ...
type LogFormatter struct{}

// Language represents collection of phrases by certain language used in application
type Language map[string]string

// Row ...
type Row struct {
	ID        int
	Word      string
	Translate string
}

// User represents user model
type User struct {
	UserID   string
	Username string
	Name     string
	Mode     string
}

// State represents last user state by saving question and answers in db.
// When user answers we get last state and compare text user send with state answer
type State struct {
	UserKey   string
	Question  Row
	Answers   []Row
	Mode      string
	Timestamp int64
}

// DBService ...
type DBService interface {
	RandomQuestion() (Row, error)
	RandomAnswers(q Row, limit int) ([]Row, error)
	SetState(State) error
	GetState(string) (State, error)
	ResetState(string) error
	InsertWord(Row) error
	InsertUser(user User) error
	UpdateUser(user User) error
	ExistUser(user User) (bool, error)
	GetUser(userid string) (User, error)
	SetUserMode(userID string, mode string) error
}

// Format ...
func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	t := entry.Time.Format("2006-01-02T15:04:05.999Z07:00")
	return []byte(fmt.Sprintf("[%s][%s][v1.0.0] %s\n", t, entry.Level.String(), entry.Message)), nil
}

// GetLanguage returns language object with phrases
func GetLanguage(content []byte) (Language, error) {
	var lang = Language{}
	err := json.Unmarshal(content, &lang)
	if err != nil {
		return nil, err
	}

	return lang, nil
}
