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
	lang     golearn.Lang
	langCode string
	token    string
	api      string
	cols     int
}

// New returns new instance of telegram hanlder
func New(cfg Config) *Handler {
	return &Handler{
		service:  cfg.Service,
		lang:     cfg.Lang,
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
	case h.lang[h.langCode]["main_menu"]:
		err = h.mainMenu(update)
	case h.lang[h.langCode]["help"]:
		err = h.help(update)
	case h.lang[h.langCode]["start"]:
		h.sendWelcomeMessage(update)
		err = h.start(update)
	case h.lang[h.langCode]["next_word"]:
		err = h.start(update)
	case h.lang[h.langCode]["again"]:
		err = h.again(update)
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
		h.sendMessage(update.Message.Chat.ID, h.lang[h.langCode]["no_words"], ReplyMarkup{})
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
		[]string{h.lang[h.langCode]["next_word"]},
	}
	message := h.lang[h.langCode]["right"]
	if !isRight {
		message = h.lang[h.langCode]["wrong"]
		reply.Keyboard = append(reply.Keyboard, []string{h.lang[h.langCode]["again"]})
	}

	h.sendMessage(update.Message.Chat.ID, message, reply)

	return nil
}

func (h *Handler) help(update TUpdate) error {
	var reply ReplyMarkup

	reply.ResizeKeyboard = true
	reply.Keyboard = [][]string{
		[]string{
			h.lang[h.langCode]["start"],
			h.lang[h.langCode]["help"],
		},
	}

	h.sendMessage(update.Message.Chat.ID, h.lang[h.langCode]["help_message"], reply)

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

	keyboard[rows] = []string{"/" + h.lang[h.langCode]["main_menu"]}

	reply.Keyboard = keyboard
	reply.ResizeKeyboard = true

	return reply
}

func (h *Handler) sendWelcomeMessage(update TUpdate) {
	keyboard := h.baseKeyboardCommands()

	h.sendMessage(update.Message.Chat.ID, h.lang[h.langCode]["welcome"], keyboard)
}

func (h *Handler) baseKeyboardCommands() ReplyMarkup {
	var reply ReplyMarkup

	keyboard := [][]string{
		[]string{h.lang[h.langCode]["start"], h.lang[h.langCode]["help"]},
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
	reply := ReplyMarkup{
		Keyboard: [][]string{
			[]string{h.lang[h.langCode]["start"], h.lang[h.langCode]["help"]},
		},
		ResizeKeyboard: true,
	}

	h.sendMessage(update.Message.Chat.ID, h.lang[h.langCode]["welcome"], reply)

	return nil
}
