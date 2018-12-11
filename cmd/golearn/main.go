package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/api"
	"github.com/sergeiten/golearn/kakaotalk"
	"github.com/sergeiten/golearn/mongo"
	"github.com/sergeiten/golearn/telegram"
	log "github.com/sirupsen/logrus"
)

var port = flag.Int("port", 8888, "Server port")

func init() {
	flag.Parse()

	log.SetFormatter(&golearn.LogFormatter{})
	log.SetLevel(log.Level(5))
}

func main() {
	cfg := golearn.ConfigFromEnv()
	language, err := golearn.GetLanguage("./lang.json")
	if err != nil {
		log.WithError(err).Fatal("failed to get language instance")
	}

	service, err := mongo.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to create mongodb instance")
	}

	cols, err := strconv.Atoi(os.Getenv("TELEGRAM_COLS_COUNT"))
	if err != nil {
		log.WithError(err).Errorf("failed to get telegram cols count")
		cols = 2 // default value
	}

	err = telegram.New(telegram.Config{
		Service:         service,
		Lang:            language,
		DefaultLanguage: cfg.DefaultLanguage,
		Token:           os.Getenv("TELEGRAM_BOT_TOKEN"),
		API:             os.Getenv("TELEGRAM_API_URL"),
		ColsCount:       cols,
	}).Serve()
	if err != nil {
		log.Fatalf("failed to start handler: %v", err)
	}

	err = api.New(service).Serve()
	if err != nil {
		log.WithError(err).Fatal("failed to start serving telegram handler")
	}

	err = kakaotalk.New(kakaotalk.Config{
		Service:         service,
		Lang:            language,
		DefaultLanguage: cfg.DefaultLanguage,
	}).Serve()
	if err != nil {
		log.WithError(err).Fatal("failed to start serving kakaotalk handler")
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
