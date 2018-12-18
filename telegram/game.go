package telegram

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sergeiten/golearn"
)

func (h *Handler) start(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	switch h.user.Mode {
	case golearn.ModePicking:
		return h.startWithPickingMode(update)
	case golearn.ModeTyping:
		return h.startWithTypingMode(update)
	default:
		return "", ReplyMarkup{}, errors.New(fmt.Sprintf("failed to start, undefined mode for user: %v", h.user))
	}
}

func (h *Handler) startWithPickingMode(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	question, err := h.db.RandomQuestion()
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	answers, err := h.db.RandomAnswers(question, 4)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	if len(answers) == 0 {
		return h.lang["no_words"], ReplyMarkup{}, nil
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
		UserKey:   update.UserID,
		Question:  question,
		Answers:   shuffledAnswers,
		Timestamp: time.Now().Unix(),
	}

	err = h.db.SetState(s)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	return question.Word, keyboard, nil
}

func (h *Handler) startWithTypingMode(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	question, err := h.db.RandomQuestion()
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	// save state
	s := golearn.State{
		UserKey:   update.UserID,
		Question:  question,
		Answers:   []golearn.Row{},
		Timestamp: time.Now().Unix(),
	}

	err = h.db.SetState(s)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	keyboard := ReplyMarkup{
		[][]string{
			{
				h.lang["next_word"],
				h.lang["show_answer"],
				h.lang["main_menu"],
			},
		},
		true,
	}

	return question.Translate, keyboard, nil
}

func (h *Handler) answer(update *golearn.Update) (message string, markup ReplyMarkup, err error) {

	state, err := h.db.GetState(update.UserID)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	var keyboard ReplyMarkup
	isRight := h.isAnswerRight(state, update)

	keyboard.ResizeKeyboard = true
	keyboard.Keyboard = [][]string{
		{h.lang["next_word"]},
	}
	message = h.lang["right"]
	if !isRight {
		message = h.lang["wrong"]

		if h.user.Mode == golearn.ModePicking {
			keyboard.Keyboard = append(keyboard.Keyboard, []string{h.lang["again"]})
		}

		if h.user.Mode == golearn.ModeTyping {
			message += "\n\n" + fmt.Sprintf(h.lang["right_answer_is"], state.Question.Word)
		}
	}

	return message, keyboard, nil
}

func (h *Handler) showAnswer(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	state, err := h.db.GetState(update.UserID)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	keyboard := ReplyMarkup{
		[][]string{
			{
				h.lang["next_word"],
				h.lang["main_menu"],
			},
		},
		true,
	}

	message = fmt.Sprintf(h.lang["right_answer_is"], state.Question.Word)

	return message, keyboard, nil
}

func (h *Handler) isAnswerRight(state golearn.State, update *golearn.Update) bool {
	toCompare := state.Question.Translate
	if h.user.Mode == golearn.ModeTyping {
		toCompare = state.Question.Word
	}
	return toCompare == update.Message
}

func (h *Handler) again(update *golearn.Update) (message string, markup ReplyMarkup, err error) {
	state, err := h.db.GetState(update.UserID)
	if err != nil {
		return "", ReplyMarkup{}, err
	}

	var keyboard ReplyMarkup
	keyboard.ResizeKeyboard = true

	for _, b := range state.Answers {
		keyboard.Keyboard = append(keyboard.Keyboard, []string{b.Translate})
	}

	return state.Question.Word, keyboard, nil
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
