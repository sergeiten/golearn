package telegram

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mongo"

	log "github.com/sirupsen/logrus"
)

var botToken = "644925777:AAEJyzTEOSTCyXdxutKYWTaFA-A3tTPxeTA"
var handler *Handler
var lang golearn.Lang

func init() {
	log.SetFormatter(&golearn.LogFormatter{})
	log.SetLevel(log.Level(5))

	cfg := &golearn.Config{
		Database: golearn.Database{
			Host:     "localhost",
			Port:     "27017",
			User:     "",
			Name:     "golearn",
			Password: "",
		},
		DefaultLanguage: "ru",
	}
	lang = golearn.GetLanguage("../lang.json")

	service, err := mongo.New(cfg)
	if err != nil {
		log.WithError(err).Fatalf("failed to create mongodb instance: %v", err)
	}

	handler = New(Config{
		Service:         service,
		Lang:            lang,
		DefaultLanguage: cfg.DefaultLanguage,
		Token:           botToken,
		API:             "https://api.telegram.org",
		ColsCount:       2,
	})
}

func TestHandler(t *testing.T) {
	commands := []string{
		lang["ru"]["main_menu"],
		lang["ru"]["help"],
		lang["ru"]["start"],
		lang["ru"]["next_word"],
		lang["ru"]["again"],
	}

	err := handler.Serve()
	if err != nil {
		t.Fatalf("failed to serve: %v", err)
	}

	for _, command := range commands {
		update := TUpdate{
			UpdateID: 148790442,
			Message: TMessage{
				MessageID: 27,
				Text:      command,
				Date:      1459919262,
			},
		}

		dat, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("failed to marshal update: %v", err)
		}

		req, err := http.NewRequest("POST", "/"+botToken+"/processMessage/", strings.NewReader(string(dat)))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
		}

		data, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Errorf("failed to read response body: %v", err)
		}

		resp := string(data)

		if resp != "OK" {
			t.Errorf("handler response wrong body: got %s, want %s", resp, "OK")
		}
	}
}

func TestReplyKeyboardWithAnswers(t *testing.T) {
	expectedString := `{"keyboard":[["test0","test1"],["test2","test3"],["/Главное Меню"]],"resize_keyboard":true}`

	words := []golearn.Row{
		{
			Word:      "test0",
			Translate: "test0",
		}, {
			Word:      "test1",
			Translate: "test1",
		},
		{
			Word:      "test2",
			Translate: "test2",
		}, {
			Word:      "test3",
			Translate: "test3",
		},
	}

	reply := handler.replyKeyboardWithAnswers(words)

	byt, err := json.Marshal(reply)
	if err != nil {
		t.Fatal("failed to unmarshal reply")
	}

	json := string(byt)

	if expectedString != json {
		t.Errorf("reply keyboard failed, expected: %s, got: %s", expectedString, json)
	}
}
