package telegram

import "github.com/sergeiten/golearn"

// TUpdate ...
type TUpdate struct {
	UpdateID int      `json:"update_id"`
	Message  TMessage `json:"message"`
}

// TMessage ...
type TMessage struct {
	MessageID int    `json:"message_id"`
	Chat      TChat  `json:"chat"`
	Text      string `json:"text"`
	Date      int    `json:"date"`
}

// TChat ...
type TChat struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	ID        int    `json:"id"`
}

// ReplyMarkup ...
type ReplyMarkup struct {
	Keyboard       [][]string `json:"keyboard"`
	ResizeKeyboard bool       `json:"resize_keyboard"`
}

// Config ...
type Config struct {
	Token           string
	ColsCount       int
	API             string
	Service         golearn.DBService
	Lang            golearn.Lang
	DefaultLanguage string
}
