package mongo

import (
	"log"
	"testing"
	"time"

	"github.com/sergeiten/golearn"
	"github.com/stretchr/testify/assert"
)

var dbService *Service
var testUser golearn.User
var testWords []golearn.Row
var testState golearn.State
var testActivities []golearn.Activity

func prepare(seeding bool) error {
	cfg := &golearn.Config{}

	cfg.Database.Host = "127.0.0.1"
	cfg.Database.Port = "27017"
	cfg.Database.Name = "test"
	cfg.DefaultLanguage = "ru"

	var err error

	dbService, err = New(cfg)
	if err != nil {
		return err
	}

	testUser = golearn.User{
		UserID:   "177374215",
		Username: "sergeiten",
		Name:     "Sergei",
		Mode:     golearn.ModePicking,
		Category: "category",
	}

	testWords = []golearn.Row{
		{
			Word:      "origin word 1",
			Translate: "translated word 1",
			Category:  "category",
		},
		{
			Word:      "origin word 2",
			Translate: "translated word 2",
			Category:  "category",
		},
		{
			Word:      "origin word 3",
			Translate: "translated word 3",
			Category:  "category",
		},
		{
			Word:      "origin word 4",
			Translate: "translated word 4",
			Category:  "category",
		},
		{
			Word:      "origin word 5",
			Translate: "translated word 5",
			Category:  "",
		},
		{
			Word:      "origin word 6",
			Translate: "translated word 6",
			Category:  "category 2",
		},
	}

	testState = golearn.State{
		UserKey:  testUser.UserID,
		Question: testWords[0],
		Answers: []golearn.Row{
			testWords[0],
			testWords[1],
			testWords[2],
			testWords[3],
		},
		Mode:      golearn.ModePicking,
		Timestamp: time.Now().Unix(),
	}

	testActivities = []golearn.Activity{
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   true,
			Timestamp: time.Date(2019, 2, 14, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   true,
			Timestamp: time.Date(2019, 2, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   true,
			Timestamp: time.Date(2019, 2, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   true,
			Timestamp: time.Date(2019, 2, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   false,
			Timestamp: time.Date(2019, 2, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   true,
			Timestamp: time.Date(2019, 2, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   false,
			Timestamp: time.Date(2019, 2, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   false,
			Timestamp: time.Date(2019, 2, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   false,
			Timestamp: time.Date(2019, 2, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   false,
			Timestamp: time.Date(2019, 2, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			UserID:    testUser.UserID,
			State:     testState,
			Answer:    "test",
			IsRight:   false,
			Timestamp: time.Date(2019, 2, 22, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := clean(); err != nil {
		return err
	}

	if seeding {
		if err := seed(); err != nil {
			return err
		}
	}

	return nil
}

func seed() error {
	err := dbService.InsertUser(testUser)

	if err != nil {
		log.Fatalf("failed to insert test user")
	}

	for _, word := range testWords {
		err := dbService.InsertWord(word)

		if err != nil {
			log.Fatalf("failed to insert test user")
		}
	}

	for _, activity := range testActivities {
		err := dbService.InsertActivity(activity)
		if err != nil {
			log.Fatalf("failed to insert activity")
		}
	}

	return nil
}

func clean() error {
	return dbService.session.DB(dbService.db).DropDatabase()
}

func TestService_RandomQuestion(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	question, err := dbService.RandomQuestion(testUser.Category)

	assert.Nil(t, err, "failed to get random question")
	assert.NotEmpty(t, question, "random question is empty")
}

func TestService_RandomAnswers(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	count := 4
	answers, err := dbService.RandomAnswers(testWords[0], count)

	assert.Nil(t, err, "failed to get random answers")
	assert.NotEmpty(t, answers)
	assert.Equal(t, count, len(answers))
	assert.Contains(t, answers, testWords[0])
}

func TestService_SetState(t *testing.T) {
	if err := prepare(false); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	err := dbService.SetState(testState)

	assert.Nil(t, err, "failed to set user state")
}

func TestService_GetState(t *testing.T) {
	if err := prepare(false); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	err := dbService.SetState(testState)

	assert.Nil(t, err)

	state, err := dbService.GetState(testUser.UserID)

	assert.Nil(t, err)
	assert.Equal(t, testState, state)
}

func TestService_InsertWord(t *testing.T) {
	if err := prepare(false); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	err := dbService.InsertWord(golearn.Row{
		Word:      "origin",
		Translate: "translate",
	})

	assert.Nil(t, err)
}

func TestService_InsertUser(t *testing.T) {
	if err := prepare(false); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	err := dbService.InsertUser(testUser)

	assert.Nil(t, err)

	user, err := dbService.GetUser(testUser.UserID)

	assert.Nil(t, err)
	assert.Equal(t, testUser, user)
}

func TestService_UpdateUser(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	updatedUser := testUser

	updatedUser.Name = "New Name"
	updatedUser.Mode = golearn.ModeTyping
	updatedUser.Category = "new category"
	updatedUser.Username = "new username"

	err := dbService.UpdateUser(updatedUser)

	assert.Nil(t, err)

	user, err := dbService.GetUser(updatedUser.UserID)

	assert.Nil(t, err)
	assert.Equal(t, updatedUser, user)
}

func TestService_ExistUser(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	exist, err := dbService.ExistUser(testUser)

	assert.Nil(t, err)
	assert.Equal(t, true, exist)
}

func TestService_GetUser(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	user, err := dbService.GetUser(testUser.UserID)

	assert.Nil(t, err, "failed to get user")
	assert.Equal(t, testUser, user)
}

func TestService_SetUserMode(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	err := dbService.SetUserMode(testUser.UserID, golearn.ModeTyping)

	assert.Nil(t, err)

	user, err := dbService.GetUser(testUser.UserID)

	assert.Nil(t, err)
	assert.Equal(t, golearn.ModeTyping, user.Mode)
}

func TestService_SetUserCategory(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	testCategory := "test category"

	err := dbService.SetUserCategory(testUser.UserID, testCategory)

	assert.Nil(t, err)

	user, err := dbService.GetUser(testUser.UserID)

	assert.Nil(t, err)
	assert.Equal(t, testCategory, user.Category)
}

func TestService_GetCategories(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	expectedCategories := []golearn.Category{
		{
			Name:  "category",
			Words: 4,
		},
		{
			Name:  "category 2",
			Words: 1,
		},
	}

	categories, err := dbService.GetCategories(testUser.UserID)

	assert.Nil(t, err)
	assert.Equal(t, len(expectedCategories), len(categories))
	assert.Equal(t, expectedCategories, categories)
}

func TestService_InsertActivity(t *testing.T) {
	if err := prepare(false); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	err := dbService.InsertActivity(golearn.Activity{
		UserID:    testUser.UserID,
		State:     testState,
		IsRight:   true,
		Answer:    "test answer",
		Timestamp: time.Now(),
	})

	assert.Nil(t, err)
}

func TestService_GetStatistics(t *testing.T) {
	if err := prepare(true); err != nil {
		t.Fatalf("failed to prepare test db: %v", err)
	}

	expectedStatistics := golearn.Statistics{
		Today: golearn.StatRow{
			Total: 3,
			Right: 0,
			Wrong: 3,
		},
		Week: golearn.StatRow{
			Total: 9,
			Right: 3,
			Wrong: 6,
		},
		Month: golearn.StatRow{
			Total: 11,
			Right: 5,
			Wrong: 6,
		},
	}

	statistics, err := dbService.GetStatistics(testUser.UserID, 2019, 2, 8, 22)

	assert.Nil(t, err)
	assert.Equal(t, expectedStatistics, statistics)
}
