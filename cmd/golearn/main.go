package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/api"
	"github.com/sergeiten/golearn/kakaotalk"
	"github.com/sergeiten/golearn/mongo"
	"github.com/sergeiten/golearn/telegram"
)

var port = flag.Int("port", 8888, "Server port")

func init() {
	flag.Parse()
}

func main() {
	cfg := golearn.ConfigFromEnv()
	langFilename := fmt.Sprintf("./lang.%s.json", cfg.DefaultLanguage)
	languageContent, err := ioutil.ReadFile(langFilename)
	golearn.LogFatal(err, "failed to get language file content")

	language, err := golearn.GetLanguage(languageContent)
	golearn.LogFatal(err, "failed to get language instance")

	service, err := mongo.New(cfg)
	golearn.LogFatal(err, "failed to create mongodb instance")
	defer service.Close()

	cols, err := strconv.Atoi(os.Getenv("TELEGRAM_COLS_COUNT"))
	if err != nil {
		golearn.LogPrint(err, "failed to get telegram cols count")
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
	golearn.LogFatal(err, "failed to start serving telegram handler")

	err = kakaotalk.New(kakaotalk.Config{
		Service:         service,
		Lang:            language,
		DefaultLanguage: cfg.DefaultLanguage,
	}).Serve()
	golearn.LogFatal(err, "failed to start serving kakaotalk handler")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
