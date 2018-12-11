package telegram

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/sergeiten/golearn"
)

func (h *Handler) start(update TUpdate) error {
	switch h.user.Mode {
	case golearn.ModePicking:
		return h.startWithPickingMode(update)
	case golearn.ModeTyping:
		return h.startWithTypingMode(update)
	default:
		log.Printf("failed to start, undefined mode for user: %v", h.user)
		h.sendWelcomeMessage(update)
	}

	return nil
}

func (h *Handler) startWithPickingMode(update TUpdate) error {
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

func (h *Handler) startWithTypingMode(update TUpdate) error {
	question, err := h.service.RandomQuestion()
	if err != nil {
		return err
	}

	// save state
	s := golearn.State{
		UserKey:   strconv.Itoa(update.Message.Chat.ID),
		Question:  question,
		Answers:   []golearn.Row{},
		Timestamp: time.Now().Unix(),
	}

	err = h.service.SetState(s)
	if err != nil {
		return err
	}

	keyboard := ReplyMarkup{
		[][]string{
			{
				h.lang["main_menu"],
				h.lang["next_word"],
			},
		},
		true,
	}

	h.sendMessage(update.Message.Chat.ID, question.Translate, keyboard)

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
		{h.lang["next_word"]},
	}
	message := h.lang["right"]
	if !isRight {
		message = h.lang["wrong"]

		if h.user.Mode == golearn.ModePicking {
			reply.Keyboard = append(reply.Keyboard, []string{h.lang["again"]})
		}

		if h.user.Mode == golearn.ModeTyping {
			message += "\n\n" + fmt.Sprintf(h.lang["right_answer_is"], state.Question.Word)
		}
	}

	h.sendMessage(update.Message.Chat.ID, message, reply)

	return nil
}

func (h *Handler) isAnswerRight(state golearn.State, update TUpdate) bool {
	toCompare := state.Question.Translate
	if h.user.Mode == golearn.ModeTyping {
		toCompare = state.Question.Word
	}
	return toCompare == update.Message.Text
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
