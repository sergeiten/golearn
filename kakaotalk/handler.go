package kakaotalk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/sergeiten/golearn"
)

// Handler ...
type Handler struct {
	service  golearn.DBService
	lang     golearn.Language
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
		golearn.LogPrint(err, "failed to unmarshal command")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	msg, err := h.prepareMessage(cmd)
	if err != nil {
		golearn.LogPrintf(err, "failed to handle message %s", cmd)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(msg)
	if err != nil {
		golearn.LogPrint(err, "failed to marshal message")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, string(resp))
}

func (h *Handler) prepareMessage(cmd *command) (*message, error) {
	switch cmd.Content {
	case h.lang["help"]:
		return h.handleHelp()
	case h.lang["start"]:
		return h.handleStart(cmd)
	case h.lang["next_word"]:
		return h.handleStart(cmd)
	case h.lang["again"]:
		return h.handleAgain(cmd)
	default:
		return h.handleCommand(cmd)
	}
}

func (h *Handler) commandFromRequest(r *http.Request) (*command, error) {
	cmd := &command{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return cmd, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(data, cmd)
	if err != nil {
		return cmd, err
	}

	return cmd, nil
}

func (h *Handler) keyboard(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(&keyboard{
		Type: typeButtons,
		Buttons: []string{
			h.lang["start"],
			h.lang["help"],
		},
	})
	if err != nil {
		golearn.LogPrint(err, "failed to marshal keyboard")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprint(w, string(resp))
}

func (h *Handler) handleHelp() (*message, error) {
	msg := &message{}

	msg.Message.Text = h.lang["help_message"]
	msg.Keyboard.Type = "buttons"
	msg.Keyboard.Buttons = []string{
		h.lang["start"],
		h.lang["help"],
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
		h.lang["next_word"],
	}
	if isRight {
		m.Message.Text = h.lang["right"]
	} else {
		m.Message.Text = h.lang["wrong"]
		m.Keyboard.Buttons = append(m.Keyboard.Buttons, h.lang["again"])
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
