package main

import (
	"math/rand"
	"time"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mongo"
)

func main() {
	cfg := &golearn.Config{
		Database: golearn.Database{
			Host: "127.0.0.1",
			Port: "27017",
			Name: "golearn",
		},
	}
	service, err := mongo.New(cfg)
	golearn.LogFatal(err, "failed to create mongodb instance")
	defer service.Close()

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

	rnd := rand.New(rand.NewSource(time.Now().Unix()))

	now := func() time.Time {
		day := rnd.Intn(31)

		return time.Date(2019, 2, day, 0, 0, 0, 0, time.UTC)
	}

	for i := 0; i < 10000; i++ {
		isRight := true
		r := rnd.Intn(2)
		if r == 1 {
			isRight = false
		}
		err = service.InsertActivity(golearn.Activity{
			UserID:    "177374215",
			State:     state,
			Answer:    "question word",
			IsRight:   isRight,
			Timestamp: now(),
		})

		if err != nil {
			golearn.LogPrint(err, "failed to insert activity")
		}
	}
}
