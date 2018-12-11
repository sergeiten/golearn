package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (h *Handler) baseKeyboardCommands() ReplyMarkup {
	var reply ReplyMarkup

	keyboard := [][]string{
		{h.lang["start"], h.lang["settings"], h.lang["help"]},
	}

	reply.Keyboard = keyboard
	reply.ResizeKeyboard = true

	return reply
}

func (h *Handler) sendWelcomeMessage(update TUpdate) {
	keyboard := h.baseKeyboardCommands()

	h.sendMessage(update.Message.Chat.ID, h.lang["welcome"], keyboard)
}

func (h *Handler) mainMenu(update TUpdate) error {
	reply := h.mainMenuKeyboard()

	h.sendMessage(update.Message.Chat.ID, h.lang["welcome"], reply)

	return nil
}

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

func (h *Handler) help(update TUpdate) error {
	var reply ReplyMarkup

	reply.ResizeKeyboard = true
	reply.Keyboard = [][]string{
		{
			h.lang["start"],
			h.lang["settings"],
			h.lang["help"],
		},
	}

	h.sendMessage(update.Message.Chat.ID, h.lang["help_message"], reply)

	return nil
}

func (h *Handler) sendMessage(chatID int, message string, reply ReplyMarkup) {
	client := &http.Client{}
	values := url.Values{}

	replyJSON, _ := json.Marshal(reply)

	values.Set("text", message)
	values.Set("chat_id", fmt.Sprintf("%d", chatID))
	values.Set("parse_mode", "HTML")
	values.Set("reply_markup", string(replyJSON))

	req, err := http.NewRequest("POST", h.api+"/bot"+h.token+"/sendMessage", strings.NewReader(values.Encode()))

	if err != nil {
		log.WithError(err).Print("failed to create request")
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)

	if err != nil {
		log.WithError(err).Print("failed to send request")
	}

	log.Debugf("POST values: %+v", values)

	defer response.Body.Close()
}

func (h *Handler) parseUpdate(r *http.Request) TUpdate {
	var update TUpdate

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("failed to get body data: %v", err)
	}

	defer r.Body.Close()

	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("failed to parse parse update json: %v", err)
	}

	return update
}
