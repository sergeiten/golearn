package telegram

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mocks"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	httpService := &mocks.HttpService{}

	update := TUpdate{
		UpdateID: 148790442,
		Message: TMessage{
			Chat: TChat{
				ID:        177374215,
				Username:  "sergeiten",
				Firstname: "Sergei",
			},
			MessageID: 27,
			Text:      "command",
			Date:      1459919262,
		},
	}

	expectedUpdate := &golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	d, _ := json.Marshal(update)

	req, _ := http.NewRequest("POST", "/", strings.NewReader(string(d)))

	httpService.On("Parse", req).Return(expectedUpdate, nil)

	u, err := httpService.Parse(req)

	assert.Equal(t, expectedUpdate, u)
	assert.Equal(t, nil, err)

	httpService.AssertExpectations(t)
}

func TestHttp_Send(t *testing.T) {
	httpService := &mocks.HttpService{}

	update := &golearn.Update{
		ChatID:   "177374215",
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Message:  "command",
	}

	httpService.On("Send", update, "", "").Return(nil)

	err := httpService.Send(update, "", "")

	assert.Equal(t, err, nil)

	httpService.AssertExpectations(t)
}
