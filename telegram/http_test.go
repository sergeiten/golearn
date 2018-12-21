package telegram

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	httpService := NewHTTP(HTTPConfig{})

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

	req := httptest.NewRequest("POST", "/", strings.NewReader(string(d)))

	u, err := httpService.Parse(req)

	assert.Equal(t, expectedUpdate, u)
	assert.Equal(t, nil, err)
}
