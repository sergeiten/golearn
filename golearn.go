package golearn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// LogFormatter ...
type LogFormatter struct{}

// Format ...
func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	t := entry.Time.Format("2006-01-02T15:04:05.999Z07:00")
	return []byte(fmt.Sprintf("[%s][%s][v1.0.0] %s\n", t, entry.Level.String(), entry.Message)), nil
}

// Phrase ...
type Phrase map[string]string

// Lang ...
type Lang map[string]Phrase

// Row ...
type Row struct {
	ID        int
	Word      string
	Translate string
}

// State ...
type State struct {
	UserKey   string
	Question  Row
	Answers   []Row
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
}

// GetLanguage returns lang object with phrases
func GetLanguage(file string) Lang {
	jsonLang, err := ioutil.ReadFile(file)
	if err != nil {
		log.WithError(err).Fatalf("failed to open language json file")
	}

	var lang = Lang{}
	err = json.Unmarshal(jsonLang, &lang)
	if err != nil {
		log.WithError(err).Fatalf("failed to unmarshal json language file")
	}

	return lang
}