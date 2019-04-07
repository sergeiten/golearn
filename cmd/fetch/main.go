package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sergeiten/golearn"
	"github.com/sergeiten/golearn/mongo"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

const (
	spreadsheetID = "1Br_w4sPm89TnkKZuCKB4J_keXFvjU_0hHRh5ro8Uvus"
)

func main() {
	ctx := context.Background()

	cfg := golearn.ConfigFromEnv()

	for {
		service, err := mongo.New(cfg)
		golearn.LogFatal(err, "failed to create mongodb instance")
		defer service.Close()

		gsrv, err := getService()
		if err != nil {
			golearn.LogFatal(err, "failed to get service")
		}

		spreadsheet, err := gsrv.Spreadsheets.Get(spreadsheetID).Do()
		if err != nil {
			golearn.LogFatal(err, "failed to get spread sheet")
		}

		categories, err := service.GetCategories("")
		if err != nil {
			golearn.LogFatal(err, "failed to get categories")
		}

		for _, sheet := range spreadsheet.Sheets {
			title := sheet.Properties.Title
			sheetRange := title + "!A:B"
			if inCategory(title, categories) {
				err := service.DeleteWordsByCategory("", title)
				if err != nil {
					golearn.LogFatal(err, "failed to delete words by category")
				}
			}
			values, err := gsrv.Spreadsheets.Values.Get(spreadsheetID, sheetRange).Context(ctx).Do()
			if err != nil {
				golearn.LogFatal(err, "failed to get spread sheet values")
				return
			}

			for _, val := range values.Values {
				w := golearn.Row{
					Word:      val[0].(string),
					Translate: val[1].(string),
					Category:  title,
				}
				err := service.InsertWord(w)
				if err != nil {
					golearn.LogFatal(err, "failed to insert word")
				}
			}
		}

		log.Printf("updated at %s\n", time.Now().Format("2006-01-02 15:04:05"))
		time.Sleep(time.Hour * 48) // sleep 2 days
	}
}

func inCategory(name string, categories []golearn.Category) bool {
	for _, category := range categories {
		if category.Name == name {
			return true
		}
	}

	return false
}

func getClient(config *oauth2.Config) (*http.Client, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

func getConfig() (*oauth2.Config, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return nil, err
	}

	return config, nil
}

func getService() (*sheets.Service, error) {
	var err error
	var config *oauth2.Config
	config, err = getConfig()
	if err != nil {
		return nil, err
	}

	var client *http.Client
	client, err = getClient(config)
	if err != nil {
		return nil, err
	}

	var srv *sheets.Service
	srv, err = sheets.New(client)
	if err != nil {
		return nil, err
	}

	return srv, err
}
