package telegram

import (
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/stretchr/testify/assert"
)

func TestMainMenuKeyboard(t *testing.T) {
	handler = New(HandlerConfig{
		DBService:       nil,
		HTTPService:     nil,
		Lang:            lang,
		DefaultLanguage: "ru",
		Token:           botToken,
		ColsCount:       2,
	})

	expectedMarkup := ReplyMarkup{
		Keyboard: [][]string{
			{
				lang["start"],
				lang["settings"],
				lang["help"],
			},
		},
		ResizeKeyboard: true,
	}

	markup := handler.mainMenuKeyboard()

	assert.Equal(t, expectedMarkup, markup)
}

func TestMainMenu(t *testing.T) {
	handler = New(HandlerConfig{
		DBService:       nil,
		HTTPService:     nil,
		Lang:            lang,
		DefaultLanguage: "ru",
		Token:           botToken,
		ColsCount:       2,
	})

	expectedMessage := lang["welcome"]
	expectedMarkup := ReplyMarkup{
		Keyboard: [][]string{
			{
				lang["start"],
				lang["settings"],
				lang["help"],
			},
		},
		ResizeKeyboard: true,
	}

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	message, markup, err := handler.mainMenu(&update)

	assert.Equal(t, expectedMessage, message)
	assert.Equal(t, expectedMarkup, markup)
	assert.Equal(t, nil, err)
}

func TestHelp(t *testing.T) {
	handler = New(HandlerConfig{
		DBService:       nil,
		HTTPService:     nil,
		Lang:            lang,
		DefaultLanguage: "ru",
		Token:           botToken,
		ColsCount:       2,
	})

	expectedMessage := lang["help_message"]
	expectedMarkup := ReplyMarkup{
		Keyboard: [][]string{
			{
				lang["start"],
				lang["settings"],
				lang["help"],
			},
		},
		ResizeKeyboard: true,
	}

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	message, markup, err := handler.help(&update)

	assert.Equal(t, expectedMessage, message)
	assert.Equal(t, expectedMarkup, markup)
	assert.Equal(t, nil, err)
}
