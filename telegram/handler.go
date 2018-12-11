package telegram

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) start(update TUpdate) error {
	question, err := h.service.RandomQuestion()
	if err != nil {
		return err
	}

	answers, err := h.service.RandomAnswers(question, 4)
	if err != nil {
		return err
	}

	if len(answers) == 0 {
		h.sendMessage(update.Message.Chat.ID, h.lang["no_words"], ReplyMarkup{})
		return nil
	}

	// shuffle answers
	shuffledAnswers := make([]golearn.Row, len(answers))
	perm := rand.Perm(len(answers))
	for i, v := range perm {
		shuffledAnswers[v] = answers[i]
	}

	keyboard := h.replyKeyboardWithAnswers(shuffledAnswers)

	// save state
	s := golearn.State{
		UserKey:   strconv.Itoa(update.Message.Chat.ID),
		Question:  question,
		Answers:   shuffledAnswers,
		Timestamp: time.Now().Unix(),
	}

	err = h.service.SetState(s)
	if err != nil {
		return err
	}

	h.sendMessage(update.Message.Chat.ID, question.Word, keyboard)

	return nil
}

func (h *Handler) answer(update TUpdate) error {
	var reply ReplyMarkup

	state, err := h.service.GetState(strconv.Itoa(update.Message.Chat.ID))
	if err != nil {
		return err
	}

	isRight := h.isAnswerRight(state, update)

	reply.ResizeKeyboard = true
	reply.Keyboard = [][]string{
		[]string{h.lang["next_word"]},
	}
	message := h.lang["right"]
	if !isRight {
		message = h.lang["wrong"]
		reply.Keyboard = append(reply.Keyboard, []string{h.lang["again"]})
	}

	h.sendMessage(update.Message.Chat.ID, message, reply)

	return nil
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

func (h *Handler) again(update TUpdate) error {
	var reply ReplyMarkup

	state, err := h.service.GetState(strconv.Itoa(update.Message.Chat.ID))
	if err != nil {
		return err
	}

	reply.ResizeKeyboard = true

	for _, b := range state.Answers {
		reply.Keyboard = append(reply.Keyboard, []string{b.Translate})
	}

	h.sendMessage(update.Message.Chat.ID, state.Question.Word, reply)

	return nil
}

func (h *Handler) replyKeyboardWithAnswers(answers []golearn.Row) ReplyMarkup {
	var reply ReplyMarkup
	var options []string

	r := float64(len(answers)) / float64(h.cols)
	rows := int(math.Ceil(r))

	keyboard := make([][]string, len(answers)-1)

	for _, a := range answers {
		options = append(options, a.Translate)
	}

	start := 0
	for i := 0; i < rows; i++ {
		if i > 0 {
			start = h.cols * i
		}

		finish := start + h.cols

		if finish > len(answers) {
			finish = len(options)
		}

		keyboard[i] = options[start:finish]
	}

	keyboard[rows] = []string{h.lang["main_menu"]}

	reply.Keyboard = keyboard
	reply.ResizeKeyboard = true

	return reply
}

func (h *Handler) sendWelcomeMessage(update TUpdate) {
	keyboard := h.baseKeyboardCommands()

	h.sendMessage(update.Message.Chat.ID, h.lang["welcome"], keyboard)
}

func (h *Handler) baseKeyboardCommands() ReplyMarkup {
	var reply ReplyMarkup

	keyboard := [][]string{
		{h.lang["start"], h.lang["settings"], h.lang["help"]},
	}

	reply.Keyboard = keyboard
	reply.ResizeKeyboard = true

	return reply
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

func (h *Handler) isAnswerRight(state golearn.State, update TUpdate) bool {
	return state.Question.Translate == update.Message.Text
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

func (h *Handler) createUser(update TUpdate) error {
	u := golearn.User{
		UserID:   strconv.Itoa(update.Message.Chat.ID),
		Username: update.Message.Chat.Username,
		Name:     update.Message.Chat.FirstName,
		Mode:     golearn.ModePicking,
	}

	exist, err := h.service.ExistUser(u)
	if err != nil {
		return err
	}

	if exist {
		return h.service.UpdateUser(u)
	}

	return h.service.InsertUser(u)
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
	userId := strconv.Itoa(update.Message.Chat.ID)

	exist, err := h.service.ExistUser(golearn.User{
		UserID: userId,
	})
	if err != nil {
		return err
	}

	if exist {
		return h.service.SetUserMode(userId, mode)
	}

	user := golearn.User{
		UserID:   strconv.Itoa(update.Message.Chat.ID),
		Username: update.Message.Chat.Username,
		Name:     update.Message.Chat.FirstName,
		Mode:     mode,
	}

	err = h.service.InsertUser(user)
	if err != nil {
		return err
	}

	reply := h.mainMenuKeyboard()

	h.sendMessage(update.Message.Chat.ID, h.lang["mode_set"], reply)

	return nil
}
