package telegram

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sergeiten/golearn"
)

type Http struct {
	api   string
	token string
}

type HttpConfig struct {
	Api   string
	Token string
}

func NewHttp(config HttpConfig) *Http {
	return &Http{
		api:   config.Api,
		token: config.Token,
	}
}

func (h *Http) Send(update *golearn.Update, message string, keyboard string) error {
	client := &http.Client{}
	values := url.Values{}

	values.Set("text", message)
	values.Set("chat_id", update.ChatID)
	values.Set("parse_mode", "HTML")
	values.Set("reply_markup", keyboard)

	req, err := http.NewRequest("POST", h.api+"/bot"+h.token+"/sendMessage", strings.NewReader(values.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	return nil
}

func (h *Http) Parse(r *http.Request) (*golearn.Update, error) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	tUpdate := TUpdate{}
	err = json.Unmarshal(body, &tUpdate)
	if err != nil {
		return nil, err
	}

	return &golearn.Update{
		ChatID:   strconv.Itoa(tUpdate.Message.Chat.ID),
		UserID:   strconv.Itoa(tUpdate.Message.Chat.ID),
		Username: tUpdate.Message.Chat.Username,
		Name:     tUpdate.Message.Chat.Firstname,
		Message:  tUpdate.Message.Text,
	}, nil
}
