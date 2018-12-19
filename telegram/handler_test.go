package telegram

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mocks"
	"github.com/stretchr/testify/assert"
)

var botToken = "644925777:AAEJyzTEOSTCyXdxutKYWTaFA-A3tTPxeTA"
var handler *Handler
var lang golearn.Language

func init() {
	langFilename := fmt.Sprintf("../lang.%s.json", "ru")
	langContent, err := ioutil.ReadFile(langFilename)
	golearn.LogFatal(err, "failed to get language file content")

	lang, err = golearn.GetLanguage(langContent)
	golearn.LogFatal(err, "failed to create language instance")
}

func TestGetOrCreateUser(t *testing.T) {
	sampleError := errors.New("sample error")

	testCases := map[string]struct {
		User        golearn.User
		Update      golearn.Update
		ReturnExist bool
		ReturnError error
	}{
		"user exists with no error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Update: golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "command",
			},
			ReturnExist: true,
			ReturnError: nil,
		},
		"user exists with error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Update: golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "command",
			},
			ReturnExist: false,
			ReturnError: sampleError,
		},
		"user no exists with no error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Update: golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "command",
			},
			ReturnExist: false,
			ReturnError: nil,
		},
		"user no exists with error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Update: golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "command",
			},
			ReturnExist: false,
			ReturnError: sampleError,
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

			dbService.On("ExistUser", tc.User).Return(tc.ReturnExist, tc.ReturnError)

			if tc.ReturnError == nil {
				if tc.ReturnExist {
					dbService.On("GetUser", tc.User.UserID).Return(tc.User, nil)
				} else {
					dbService.On("InsertUser", tc.User).Return(nil)
				}
			}

			user, err := handler.getOrCreateUser(&tc.Update)

			assert.Equal(t, tc.User, user)

			assert.Equal(t, tc.ReturnError, err)

			dbService.AssertExpectations(t)
		})
	}
}

func TestSettings(t *testing.T) {
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

	update := golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	expectedMessage := lang["mode_explain"]
	expectedMarkup := ReplyMarkup{
		Keyboard: [][]string{
			{
				lang["mode_picking"],
				lang["mode_typing"],
			},
		},
		ResizeKeyboard: true,
	}

	message, markup, err := handler.settings(&update)

	assert.Equal(t, expectedMessage, message)
	assert.Equal(t, expectedMarkup, markup)
	assert.Equal(t, nil, err)
}

func TestSetMode(t *testing.T) {
	testCases := map[string]struct {
		User    golearn.User
		Mode    string
		Message string
		Markup  ReplyMarkup
		Error   error
	}{
		"set mode with no error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Mode:    golearn.ModeTyping,
			Message: lang["mode_set"],
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["start"],
						lang["settings"],
						lang["help"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
		},
		"set mode with error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Mode:    golearn.ModeTyping,
			Message: "",
			Markup:  ReplyMarkup{},
			Error:   errors.New("sample error"),
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

			handler.user = tc.User

			dbService.On("SetUserMode", tc.User.UserID, tc.Mode).Return(tc.Error)

			message, markup, err := handler.setMode(tc.Mode)

			assert.Equal(t, tc.Message, message)
			assert.Equal(t, tc.Markup, markup)
			assert.Equal(t, tc.Error, err)

			dbService.AssertExpectations(t)
		})
	}
}

//func TestHandle(t *testing.T) {
//	testCases := []struct {
//		Command string
//		Message string
//		Markup  ReplyMarkup
//		Error   error
//	}{
//		{
//			Command: lang["main_menu"],
//			Message: lang["welcome"],
//			Markup: ReplyMarkup{
//				Keyboard: [][]string{
//					{
//						lang["start"],
//						lang["settings"],
//						lang["help"],
//					},
//				},
//				ResizeKeyboard: true,
//			},
//			Error: nil,
//		},
//	}
//	//commands := []string{
//	//lang["main_menu"],
//	//lang["help"],
//	//lang["start"],
//	//lang["next_word"],
//	//lang["again"],
//	//lang["settings"],
//	//lang["mode_picking"],
//	//lang["mode_picking"],
//	//lang["show_answer"],
//	//}
//
//	err := handler.Serve()
//	if err != nil {
//		t.Fatalf("failed to serve: %v", err)
//	}
//
//	update := TUpdate{
//		UpdateID: 148790442,
//		Message: TMessage{
//			Chat: TChat{
//				ID:        177374215,
//				Username:  "sergeiten",
//				Firstname: "Sergei",
//			},
//			MessageID: 27,
//			Text:      "",
//			Date:      1459919262,
//		},
//	}
//
//	for _, command := range commands {
//		update.Message.Text = command
//
//		dat, err := json.Marshal(update)
//		if err != nil {
//			t.Fatalf("failed to marshal update: %v", err)
//		}
//
//		req := httptest.NewRequest("POST", "/"+botToken+"/processMessage/", strings.NewReader(string(dat)))
//
//		rr := httptest.NewRecorder()
//
//		handler.ServeHTTP(rr, req)
//
//		//if status := rr.Code; status != http.StatusOK {
//		//	t.Errorf("handler returned wrong status code: got %v, want %v", status, http.StatusOK)
//		//}
//
//		assert.Equal(t, http.StatusOK, rr.Code)
//
//		data, err := ioutil.ReadAll(rr.Body)
//		if err != nil {
//			t.Errorf("failed to read response body: %v", err)
//		}
//
//		resp := string(data)
//
//		assert.Equal(t, "OK", resp)
//	}
//}
