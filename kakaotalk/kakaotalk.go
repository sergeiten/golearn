package kakaotalk

import "github.com/sergeiten/golearn"

const typeButtons = "buttons"

type keyboard struct {
	Type    string   `json:"type"`
	Buttons []string `json:"buttons,omitempty"`
}

type command struct {
	UserKey string `json:"user_key"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type message struct {
	Message struct {
		Text string `json:"text,omitempty"`
	} `json:"message"`
	Keyboard keyboard `json:"keyboard,omitempty"`
}

// Config ...
type Config struct {
	Token           string
	ColsCount       int
	API             string
	Service         golearn.DBService
	Lang            golearn.Language
	DefaultLanguage string
}
