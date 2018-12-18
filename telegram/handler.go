package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sergeiten/golearn"
)

// Handler telegram HTTP handler
type Handler struct {
	db       golearn.DBService
	http     golearn.HttpService
	lang     golearn.Language
	langCode string
	cols     int
	token    string
	user     golearn.User
}

// HandlerConfig handler config
type HandlerConfig struct {
	DBService       golearn.DBService
	HttpService     golearn.HttpService
	Lang            golearn.Language
	DefaultLanguage string
	Token           string
	ColsCount       int
}

// New returns new instance of telegram hanlder
func New(cfg HandlerConfig) *Handler {
	return &Handler{
		db:       cfg.DBService,
		http:     cfg.HttpService,
		lang:     cfg.Lang,
		langCode: cfg.DefaultLanguage,
		cols:     cfg.ColsCount,
		token:    cfg.Token,
	}
}

// Serve starts http handler
func (h *Handler) Serve() error {
	http.Handle("/"+h.token+"/processMessage/", h)

	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var message string
	var keyboard ReplyMarkup

	update, err := h.http.Parse(r)
	golearn.LogPrint(err, "failed to parse update")

	h.user, err = h.getOrCreateUser(update)
	if err != nil {
		golearn.LogPrint(err, "failed to get/create user")
		return
	}

	switch update.Message {
	case h.lang["main_menu"]:
		message, keyboard, err = h.main(update)
	case h.lang["help"]:
		message, keyboard, err = h.help(update)
	case h.lang["start"]:
		message, keyboard, err = h.start(update)
	case h.lang["next_word"]:
		message, keyboard, err = h.start(update)
	case h.lang["again"]:
		message, keyboard, err = h.again(update)
	case h.lang["settings"]:
		message, keyboard, err = h.settings(update)
	case h.lang["mode_picking"]:
		message, keyboard, err = h.setMode(golearn.ModePicking)
	case h.lang["mode_typing"]:
		message, keyboard, err = h.setMode(golearn.ModeTyping)
	case h.lang["show_answer"]:
		message, keyboard, err = h.showAnswer(update)
	default:
		message, keyboard, err = h.answer(update)
	}

	if err != nil {
		golearn.LogPrintf(err, "failed to handle %s command", update.Message)
		_, err = fmt.Fprint(w, err.Error())
		golearn.LogPrint(err, "failed to send response")
		return
	}

	d, err := json.Marshal(keyboard)
	if err != nil {
		golearn.LogPrint(err, "failed to marshal reply keyboard")
		return
	}

	err = h.http.Send(update, message, string(d))
	if err != nil {
		golearn.LogPrint(err, "failed to send message")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "plain/text")
	_, err = fmt.Fprint(w, "OK")
	golearn.LogPrint(err, "failed to send response")
}

func (h *Handler) getOrCreateUser(update *golearn.Update) (golearn.User, error) {
	u := golearn.User{
		UserID:   update.UserID,
		Username: update.Username,
		Name:     update.Name,
		Mode:     golearn.ModePicking,
	}

	exist, err := h.db.ExistUser(u)
	if err != nil {
		return u, err
	}

	if exist {
		return h.db.GetUser(u.UserID)
	}

	return u, h.db.InsertUser(u)
}

func (h *Handler) settings(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	keyboard := ReplyMarkup{
		Keyboard: [][]string{
			{
				h.lang["mode_picking"],
				h.lang["mode_typing"],
			},
		},
		ResizeKeyboard: true,
	}

	return h.lang["mode_explain"], keyboard, nil
}

func (h *Handler) setMode(mode string) (message string, markup ReplyMarkup, err error) {
	err = h.db.SetUserMode(h.user.UserID, mode)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	keyboard := h.mainMenuKeyboard()

	return h.lang["mode_set"], keyboard, nil
}
