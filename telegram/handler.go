package telegram

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/sergeiten/golearn"
	log "github.com/sirupsen/logrus"
)

// Handler telegram HTTP handler
type Handler struct {
	service  golearn.DBService
	lang     golearn.Phrase
	langCode string
	token    string
	api      string
	cols     int
	user     golearn.User
}

// New returns new instance of telegram hanlder
func New(cfg Config) *Handler {
	return &Handler{
		service:  cfg.Service,
		lang:     cfg.Lang[cfg.DefaultLanguage],
		langCode: cfg.DefaultLanguage,
		token:    cfg.Token,
		api:      cfg.API,
		cols:     cfg.ColsCount,
	}
}

// Serve starts http handler
func (h *Handler) Serve() error {
	http.Handle("/"+h.token+"/processMessage/", h)

	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Received %s %s", r.Method, html.EscapeString(r.URL.Path))

	var err error

	update := h.parseUpdate(r)

	h.user, err = h.getOrCreateUser(update)
	if err != nil {
		log.Printf("failed to get/create user: %v", err)
		return
	}

	switch update.Message.Text {
	case h.lang["main_menu"]:
		err = h.mainMenu(update)
	case h.lang["help"]:
		err = h.help(update)
	case h.lang["start"]:
		h.sendWelcomeMessage(update)
		err = h.start(update)
	case h.lang["next_word"]:
		err = h.start(update)
	case h.lang["again"]:
		err = h.again(update)
	case h.lang["settings"]:
		err = h.settings(update)
	case h.lang["mode_picking"]:
		err = h.setMode(update, golearn.ModePicking)
	case h.lang["mode_typing"]:
		err = h.setMode(update, golearn.ModeTyping)
	default:
		err = h.answer(update)
	}

	if err != nil {
		log.Printf("failed to make %s command: %v", update.Message.Text, err)
		_, err = fmt.Fprint(w, err.Error())
		if err != nil {
			log.Printf("failed to send response: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "plain/text")
	_, err = fmt.Fprint(w, "OK")
	if err != nil {
		log.Printf("failed to send response: %v", err)
	}
}

func (h *Handler) getOrCreateUser(update TUpdate) (golearn.User, error) {
	u := golearn.User{
		UserID:   strconv.Itoa(update.Message.Chat.ID),
		Username: update.Message.Chat.Username,
		Name:     update.Message.Chat.Firstname,
		Mode:     golearn.ModePicking,
	}

	exist, err := h.service.ExistUser(u)
	if err != nil {
		return u, err
	}

	if exist {
		return h.service.GetUser(u.UserID)
	}

	return u, h.service.InsertUser(u)
}

func (h *Handler) settings(update TUpdate) error {
	reply := ReplyMarkup{
		Keyboard: [][]string{
			{
				h.lang["mode_picking"],
				h.lang["mode_typing"],
			},
		},
		ResizeKeyboard: true,
	}

	h.sendMessage(update.Message.Chat.ID, h.lang["mode_explain"], reply)

	return nil
}

func (h *Handler) setMode(update TUpdate, mode string) error {
	err := h.service.SetUserMode(h.user.UserID, mode)
	if err != nil {
		return err
	}

	reply := h.mainMenuKeyboard()

	h.sendMessage(update.Message.Chat.ID, h.lang["mode_set"], reply)

	return nil
}
