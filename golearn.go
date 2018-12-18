package golearn

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
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

// Update represents joint response data model from service (telegram, kakaotalk).
// Contains minimal required data for detecting user and sending message back.
type Update struct {
	ChatID   string
	UserID   string
	Username string
	Name     string
	Message  string
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
	Close()
}

// HTTPService represents interface for dealing with sending and parsing http requests
type HTTPService interface {
	Send(update *Update, message string, keyboard string) error
	Parse(r *http.Request) (*Update, error)
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

// LogPrint prints error message with stack trace without exited program.
func LogPrint(err error, message string) {
	if err != nil {
		log.Printf("%+v", errors.Wrap(err, message))
	}
}

// LogPrintf prints error message with stach trace.
// Arguments are handled in the manner of fmt.Printf.
func LogPrintf(err error, message string, args ...interface{}) {
	if err != nil {
		log.Printf("%+v", errors.Wrapf(err, message, args...))
	}
}

// LogFatal prints error message with stack trace with exited program.
func LogFatal(err error, message string) {
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, message))
	}
}
