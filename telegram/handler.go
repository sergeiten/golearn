package telegram

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/sergeiten/golearn"
)

// Handler telegram HTTP handler
type Handler struct {
	db       golearn.DBService
	http     golearn.HTTPService
	lang     golearn.Language
	langCode string
	cols     int
	token    string
	user     golearn.User
}

// HandlerConfig handler config
type HandlerConfig struct {
	DBService       golearn.DBService
	HTTPService     golearn.HTTPService
	Lang            golearn.Language
	DefaultLanguage string
	Token           string
	ColsCount       int
}

// New returns new instance of telegram handler
func New(cfg HandlerConfig) *Handler {
	return &Handler{
		db:       cfg.DBService,
		http:     cfg.HTTPService,
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
	update, err := h.http.Parse(r)
	golearn.LogPrint(err, "failed to parse update")

	h.user, err = h.getOrCreateUser(update)
	if err != nil {
		golearn.LogPrint(err, "failed to get/create user")
		return
	}

	message, keyboard, err := h.handle(update)

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

func (h *Handler) handle(update *golearn.Update) (string, ReplyMarkup, error) {
	switch {
	case update.Message == h.lang["main_menu"]:
		return h.mainMenu(update)
	case update.Message == "/start":
		return h.mainMenu(update)
	case update.Message == h.lang["help"]:
		return h.help(update)
	case update.Message == h.lang["start"]:
		return h.start(update)
	case update.Message == h.lang["next_word"]:
		return h.start(update)
	case update.Message == h.lang["again"]:
		return h.again(update)
	case update.Message == h.lang["settings"]:
		return h.settings(update)
	case update.Message == h.lang["mode_picking"]:
		return h.setMode(golearn.ModePicking)
	case update.Message == h.lang["mode_typing"]:
		return h.setMode(golearn.ModeTyping)
	case update.Message == h.lang["show_answer"]:
		return h.showAnswer(update)
	case update.Message == h.lang["categories"]:
		return h.categories(update)
	case strings.HasPrefix(update.Message, h.lang["categories_icon"]):
		return h.setCategory(update)
	case update.Message == h.lang["reset_category"]:
		return h.setCategory(update)
	default:
		return h.answer(update)
	}
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
				h.lang["categories"],
			},
		},
		ResizeKeyboard: true,
	}

	return h.lang["mode_explain"], keyboard, nil
}

func (h *Handler) categories(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	categories, err := h.db.GetCategories(update.UserID)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	var reply ReplyMarkup
	var options []string

	r := float64(len(categories)) / float64(h.cols)
	rows := int(math.Ceil(r))

	keyboard := make([][]string, rows+1)

	for _, a := range categories {
		options = append(options, h.lang["categories_icon"]+" "+a.Name)
	}

	start := 0
	for i := 0; i < rows; i++ {
		if i > 0 {
			start = h.cols * i
		}

		finish := start + h.cols

		if finish > len(categories) {
			finish = len(options)
		}

		keyboard[i] = options[start:finish]
	}

	keyboard[rows] = []string{
		h.lang["reset_category"],
		h.lang["main_menu"],
	}

	reply.Keyboard = keyboard
	reply.ResizeKeyboard = true

	return h.lang["pick_category"], reply, nil
}

func (h *Handler) setCategory(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	// remove category icon
	category := strings.Trim(strings.Replace(update.Message, h.lang["categories_icon"], "", -1), " ")
	err = h.db.SetUserCategory(h.user.UserID, category)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	keyboard := h.mainMenuKeyboard()

	return h.lang["category_set"], keyboard, nil
}

func (h *Handler) setMode(mode string) (message string, markup ReplyMarkup, err error) {
	err = h.db.SetUserMode(h.user.UserID, mode)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	keyboard := h.mainMenuKeyboard()

	return h.lang["mode_set"], keyboard, nil
}
