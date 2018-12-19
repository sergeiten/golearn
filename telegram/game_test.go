package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mocks"
	"github.com/stretchr/testify/assert"
)

func TestReplyKeyboardWithAnswers(t *testing.T) {
	handler = New(HandlerConfig{
		DBService:       nil,
		HTTPService:     nil,
		Lang:            lang,
		DefaultLanguage: "ru",
		Token:           botToken,
		ColsCount:       2,
	})

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

func TestAgain(t *testing.T) {
	sampleError := errors.New("sample error")

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	testCases := map[string]struct {
		State   golearn.State
		Error   error
		Markup  ReplyMarkup
		Message string
	}{
		"again with no error": {
			State: golearn.State{
				UserKey: "177374215",
				Question: golearn.Row{
					Word:      "question word",
					Translate: "question translate",
				},
				Answers: []golearn.Row{
					{
						Word:      "answer word 1",
						Translate: "answer translate 1",
					},
					{
						Word:      "answer word 2",
						Translate: "answer translate 2",
					},
					{
						Word:      "answer word 3",
						Translate: "answer translate 3",
					},
					{
						Word:      "answer word 4",
						Translate: "answer translate 4",
					},
				},
				Mode: golearn.ModePicking,
			},
			Error: nil,
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						"answer translate 1",
					},
					{
						"answer translate 2",
					},
					{
						"answer translate 3",
					},
					{
						"answer translate 4",
					},
				},
				ResizeKeyboard: true,
			},
			Message: "question word",
		},
		"again with error": {
			State: golearn.State{
				UserKey: "177374215",
				Question: golearn.Row{
					Word:      "question word",
					Translate: "question translate",
				},
				Answers: []golearn.Row{
					{
						Word:      "answer word 1",
						Translate: "answer translate 1",
					},
					{
						Word:      "answer word 2",
						Translate: "answer translate 2",
					},
					{
						Word:      "answer word 3",
						Translate: "answer translate 3",
					},
					{
						Word:      "answer word 4",
						Translate: "answer translate 4",
					},
				},
				Mode: golearn.ModePicking,
			},
			Error: sampleError,
			Markup: ReplyMarkup{
				Keyboard:       nil,
				ResizeKeyboard: false,
			},
			Message: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dbService := &mocks.DBService{}
			httpService := &mocks.HttpService{}

			handler = New(HandlerConfig{
				DBService:       dbService,
				HTTPService:     httpService,
				Lang:            lang,
				DefaultLanguage: "ru",
				Token:           botToken,
				ColsCount:       2,
			})

			dbService.On("GetState", update.UserID).Return(tc.State, tc.Error)

			//state, err := dbService.GetState(update.UserID)

			message, markup, err := handler.again(&update)

			assert.Equal(t, tc.Message, message)
			assert.Equal(t, tc.Markup, markup)
			assert.Equal(t, tc.Error, err)

			dbService.AssertExpectations(t)
		})
	}
}

func TestIsAnswerRight(t *testing.T) {
	handler = New(HandlerConfig{
		DBService:       nil,
		HTTPService:     nil,
		Lang:            lang,
		DefaultLanguage: "ru",
		Token:           botToken,
		ColsCount:       2,
	})

	user := golearn.User{
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Mode:     golearn.ModePicking,
	}

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	testCases := map[string]struct {
		Message  string
		Mode     string
		Expected bool
		State    golearn.State
	}{
		"typing mode right answer": {
			Message:  "question word",
			Mode:     golearn.ModeTyping,
			Expected: true,
			State: golearn.State{
				Question: golearn.Row{
					Word:      "question word",
					Translate: "question translate",
				},
			},
		},
		"typing mode wrong answer": {
			Message:  "wrong",
			Mode:     golearn.ModeTyping,
			Expected: false,
			State: golearn.State{
				Question: golearn.Row{
					Word:      "question word",
					Translate: "question translate",
				},
			},
		},
		"picking mode right answer": {
			Message:  "question translate",
			Mode:     golearn.ModePicking,
			Expected: true,
			State: golearn.State{
				Question: golearn.Row{
					Word:      "question word",
					Translate: "question translate",
				},
			},
		},
		"picking mode wrong answer": {
			Message:  "wrong",
			Mode:     golearn.ModePicking,
			Expected: false,
			State: golearn.State{
				Question: golearn.Row{
					Word:      "question word",
					Translate: "question translate",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			update.Message = tc.Message
			user.Mode = tc.Mode

			handler.user = user

			isRight := handler.isAnswerRight(tc.State, &update)

			assert.Equal(t, tc.Expected, isRight)
		})
	}
}

func TestShowAnswer(t *testing.T) {
	sampleError := errors.New("sample error")

	state := golearn.State{
		UserKey: "177374215",
		Question: golearn.Row{
			Word:      "question word",
			Translate: "question translate",
		},
		Answers: []golearn.Row{
			{
				Word:      "answer word 1",
				Translate: "answer translate 1",
			},
			{
				Word:      "answer word 2",
				Translate: "answer translate 2",
			},
			{
				Word:      "answer word 3",
				Translate: "answer translate 3",
			},
			{
				Word:      "answer word 4",
				Translate: "answer translate 4",
			},
		},
		Mode: golearn.ModePicking,
	}

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	testCases := map[string]struct {
		Message string
		Markup  ReplyMarkup
		Error   error
	}{
		"with no error": {
			Message: fmt.Sprintf(lang["right_answer_is"], state.Question.Word),
			Markup: ReplyMarkup{
				[][]string{
					{
						lang["next_word"],
						lang["main_menu"],
					},
				},
				true,
			},
			Error: nil,
		},
		"with error": {
			Message: "",
			Markup:  ReplyMarkup{},
			Error:   sampleError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dbService := &mocks.DBService{}
			httpService := &mocks.HttpService{}

			handler = New(HandlerConfig{
				DBService:       dbService,
				HTTPService:     httpService,
				Lang:            lang,
				DefaultLanguage: "ru",
				Token:           botToken,
				ColsCount:       2,
			})

			dbService.On("GetState", update.UserID).Return(state, tc.Error)

			message, markup, err := handler.showAnswer(&update)

			assert.Equal(t, tc.Message, message)
			assert.Equal(t, tc.Markup, markup)
			assert.Equal(t, tc.Error, err)

			dbService.AssertExpectations(t)
		})
	}
}

func TestAnswer(t *testing.T) {
	state := golearn.State{
		UserKey: "177374215",
		Question: golearn.Row{
			Word:      "question word",
			Translate: "question translate",
		},
		Answers: []golearn.Row{
			{
				Word:      "answer word 1",
				Translate: "answer translate 1",
			},
			{
				Word:      "answer word 2",
				Translate: "answer translate 2",
			},
			{
				Word:      "answer word 3",
				Translate: "answer translate 3",
			},
			{
				Word:      "answer word 4",
				Translate: "answer translate 4",
			},
		},
		Mode: golearn.ModePicking,
	}

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	testCases := map[string]struct {
		UpdateMessage string
		Message       string
		Markup        ReplyMarkup
		Error         error
		Mode          string
	}{
		"with error": {
			UpdateMessage: "message",
			Message:       "",
			Markup:        ReplyMarkup{},
			Error:         errors.New("sample error"),
			Mode:          golearn.ModePicking,
		},
		"picking mode right answer": {
			UpdateMessage: "question translate",
			Message:       lang["right"],
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["next_word"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
			Mode:  golearn.ModePicking,
		},
		"picking mode wrong answer": {
			UpdateMessage: "wrong",
			Message:       lang["wrong"],
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["next_word"],
					},
					{
						lang["again"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
			Mode:  golearn.ModePicking,
		},
		"typing mode right answer": {
			UpdateMessage: "question word",
			Message:       lang["right"],
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["next_word"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
			Mode:  golearn.ModeTyping,
		},
		"typing mode wrong answer": {
			UpdateMessage: "wrong",
			Message:       lang["wrong"] + "\n\n" + fmt.Sprintf(lang["right_answer_is"], state.Question.Word),
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["next_word"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
			Mode:  golearn.ModeTyping,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dbService := &mocks.DBService{}
			httpService := &mocks.HttpService{}

			handler = New(HandlerConfig{
				DBService:       dbService,
				HTTPService:     httpService,
				Lang:            lang,
				DefaultLanguage: "ru",
				Token:           botToken,
				ColsCount:       2,
			})

			handler.user.Mode = tc.Mode
			update.Message = tc.UpdateMessage

			dbService.On("GetState", update.UserID).Return(state, tc.Error)

			message, markup, err := handler.answer(&update)

			assert.Equal(t, tc.Message, message)
			assert.Equal(t, tc.Markup, markup)
			assert.Equal(t, tc.Error, err)

			dbService.AssertExpectations(t)
		})
	}
}