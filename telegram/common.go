package telegram

import (
	"github.com/sergeiten/golearn"
)

func (h *Handler) mainMenuKeyboard() ReplyMarkup {
	return ReplyMarkup{
		Keyboard: [][]string{
			{
				h.lang["start"],
				h.lang["settings"],
				h.lang["help"],
			},
		},
		ResizeKeyboard: true,
	}
}

func (h *Handler) mainMenu(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	keyboard := h.mainMenuKeyboard()

	return h.lang["welcome"], keyboard, nil
}

func (h *Handler) help(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	keyboard := h.mainMenuKeyboard()

	return h.lang["help_message"], keyboard, nil
}
