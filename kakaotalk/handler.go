package kakaotalk

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/sergeiten/golearn"

	log "github.com/sirupsen/logrus"
)

// Handler ...
type Handler struct {
	service  golearn.DBService
	lang     golearn.Lang
	langCode string
}

// New returns new instance of Handler
func New(cfg Config) *Handler {
	return &Handler{
		service:  cfg.Service,
		lang:     cfg.Lang,
		langCode: cfg.DefaultLanguage,
	}
}

// Serve starts serving http
func (h *Handler) Serve() error {
	http.Handle("/kakaobot/", h)
	return nil
}

// HandleHTTP handle http requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Received %s %s\n", r.Method, html.EscapeString(r.URL.Path))

	if r.Method == http.MethodGet && r.URL.Path == "/kakaobot/keyboard" {
		h.keyboard(w, r)
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/kakaobot/message" {
		h.message(w, r)
		return
	}
}

func (h *Handler) message(w http.ResponseWriter, r *http.Request) {
	cmd, err := h.commandFromRequest(r)
	if err != nil {
		log.WithError(err).Errorf("failed to unmarshal command")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Debugf("Get command from %s user with content %s\n", cmd.UserKey, cmd.Content)

	msg, err := h.prepareMessage(cmd)
	if err != nil {
		log.WithError(err).Errorf("failed to handle message %+v: %v", cmd, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(msg)
	if err != nil {
		log.WithError(err).Errorf("failed to marshal message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, string(resp))
}

func (h *Handler) prepareMessage(cmd *command) (*message, error) {
	switch cmd.Content {
	case h.lang[h.langCode]["help"]:
		return h.handleHelp()
	case h.lang[h.langCode]["start"]:
		return h.handleStart(cmd)
	case h.lang[h.langCode]["next_word"]:
		return h.handleStart(cmd)
	case h.lang[h.langCode]["again"]:
		return h.handleAgain(cmd)
	default:
		return h.handleCommand(cmd)
	}
}

func (h *Handler) commandFromRequest(r *http.Request) (*command, error) {
	cmd := &command{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Errorf("failed to read response body")
		return cmd, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(data, cmd)
	if err != nil {
		log.WithError(err).Errorf("failed to unmarshal command")
		return cmd, err
	}

	return cmd, nil
}

func (h *Handler) keyboard(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(&keyboard{
		Type: typeButtons,
		Buttons: []string{
			h.lang[h.langCode]["start"],
			h.lang[h.langCode]["help"],
		},
	})
	if err != nil {
		log.WithError(err).Errorf("failed to marshal keyboard")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprint(w, string(resp))
}

func (h *Handler) handleHelp() (*message, error) {
	msg := &message{}

	msg.Message.Text = h.lang[h.langCode]["help_message"]
	msg.Keyboard.Type = "buttons"
	msg.Keyboard.Buttons = []string{
		h.lang[h.langCode]["start"],
		h.lang[h.langCode]["help"],
	}

	return msg, nil
}

// handleStart returns message for start action
func (h *Handler) handleStart(cmd *command) (*message, error) {
	m := &message{}
	question, err := h.service.RandomQuestion()
	if err != nil {
		return m, err
	}

	answers, err := h.service.RandomAnswers(question, 4)
	if err != nil {
		return m, err
	}

	// shuffle answers
	shuffledAnswers := make([]golearn.Row, len(answers))
	perm := rand.Perm(len(answers))
	for i, v := range perm {
		shuffledAnswers[v] = answers[i]
	}

	// prepare return message
	m.Message.Text = question.Word
	m.Keyboard.Type = "buttons"

	for _, a := range shuffledAnswers {
		m.Keyboard.Buttons = append(m.Keyboard.Buttons, a.Translate)
	}

	// save state
	s := golearn.State{
		UserKey:   cmd.UserKey,
		Question:  question,
		Answers:   shuffledAnswers,
		Timestamp: time.Now().Unix(),
	}

	err = h.service.SetState(s)
	return m, err
}

func (h *Handler) handleCommand(cmd *command) (*message, error) {
	m := &message{}
	state, err := h.service.GetState(cmd.UserKey)
	if err != nil {
		return m, err
	}

	isRight := h.isAnswerRight(state, cmd)

	m.Keyboard.Type = typeButtons
	m.Keyboard.Buttons = []string{
		h.lang[h.langCode]["next_word"],
	}
	if isRight {
		m.Message.Text = h.lang[h.langCode]["right"]
	} else {
		m.Message.Text = h.lang[h.langCode]["wrong"]
		m.Keyboard.Buttons = append(m.Keyboard.Buttons, h.lang[h.langCode]["again"])
	}

	return m, nil
}

func (h *Handler) handleAgain(cmd *command) (*message, error) {
	m := &message{}
	state, err := h.service.GetState(cmd.UserKey)
	if err != nil {
		return m, err
	}
	m.Message.Text = state.Question.Word
	m.Keyboard.Type = typeButtons

	for _, b := range state.Answers {
		m.Keyboard.Buttons = append(m.Keyboard.Buttons, b.Translate)
	}

	return m, nil
}

func (h *Handler) isAnswerRight(state golearn.State, cmd *command) bool {
	return state.Question.Translate == cmd.Content
}
