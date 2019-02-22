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
				lang["categories"],
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
						lang["statistics"],
					},
					{
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

func TestSetCategory(t *testing.T) {
	testCases := map[string]struct {
		User     golearn.User
		Update   *golearn.Update
		Category string
		Message  string
		Markup   ReplyMarkup
		Error    error
	}{
		"set category with no error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Update: &golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "2019-01-01",
			},
			Category: "2019-01-01",
			Message:  lang["category_set"],
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["start"],
						lang["statistics"],
					},
					{
						lang["settings"],
						lang["help"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
		},
		"set category with error": {
			User: golearn.User{
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Mode:     golearn.ModePicking,
			},
			Update: &golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "2019-01-01",
			},
			Category: "2019-01-01",
			Message:  "",
			Markup:   ReplyMarkup{},
			Error:    errors.New("sample error"),
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

			dbService.On("SetUserCategory", tc.User.UserID, tc.Category).Return(tc.Error)

			message, markup, err := handler.setCategory(tc.Update)

			assert.Equal(t, tc.Message, message)
			assert.Equal(t, tc.Markup, markup)
			assert.Equal(t, tc.Error, err)

			dbService.AssertExpectations(t)
		})
	}
}

func TestCategories(t *testing.T) {
	testCases := map[string]struct {
		Update     *golearn.Update
		Categories []golearn.Category
		Message    string
		Markup     ReplyMarkup
		Error      error
	}{
		"categories with no error": {
			Update: &golearn.Update{
				ChatID:   "177374215",
				UserID:   "177374215",
				Username: "sergeiten",
				Name:     "Sergei",
				Message:  "",
			},
			Categories: []golearn.Category{
				{
					Name:  "2019-01-01",
					Words: 10,
				},
				{
					Name:  "2019-01-02",
					Words: 15,
				},
			},
			Message: lang["pick_category"],
			Markup: ReplyMarkup{
				Keyboard: [][]string{
					{
						lang["categories_icon"] + " 2019-01-01",
						lang["categories_icon"] + " 2019-01-02",
					},
					{
						lang["reset_category"],
						lang["main_menu"],
					},
				},
				ResizeKeyboard: true,
			},
			Error: nil,
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

			dbService.On("GetCategories", tc.Update.UserID).Return(tc.Categories, tc.Error)

			message, markup, err := handler.categories(tc.Update)

			assert.Equal(t, tc.Message, message)
			assert.Equal(t, tc.Markup, markup)
			assert.Equal(t, tc.Error, err)

			dbService.AssertExpectations(t)
		})
	}
}
