package mongo

import (
	"testing"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_ExistUser(t *testing.T) {
	dbService := &mocks.DBService{}

	//sampleError := errors.New("sample error")

	testCases := []struct {
		User     golearn.User
		Expected bool
		Error    error
	}{
		{
			User: golearn.User{
				UserID: "177374215",
			},
			Expected: true,
			Error:    nil,
		},
		{
			User: golearn.User{
				UserID: "",
			},
			Expected: true,
			Error:    nil,
		},
	}

	for _, tc := range testCases {
		dbService.On("ExistUser", tc.User).Return(tc.Expected, tc.Error)

		exist, err := dbService.ExistUser(tc.User)

		assert.Equal(t, tc.Expected, exist)
		assert.Equal(t, tc.Error, err)

		dbService.AssertExpectations(t)
	}
}
