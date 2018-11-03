package api

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/sergeiten/golearn"
)

// API ...
type API struct {
	Service golearn.DBService
}

// New returns new api handler instance
func New(service golearn.DBService) *API {
	return &API{
		Service: service,
	}
}

// Serve starts serving http requests
func (h API) Serve() error {
	http.Handle("/api/", h)
	return nil
}

func (h API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("received %s %s\n", r.Method, html.EscapeString(r.URL.Path))

	if r.Method == http.MethodPost && r.URL.Path == "/api/word" {
		h.insertWord(w, r)
		return
	}
}

func (h API) insertWord(w http.ResponseWriter, r *http.Request) {
	word := r.FormValue("word")
	translate := r.FormValue("translate")

	if word == "" {
		log.Printf("Failed to insert word: word is empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if translate == "" {
		log.Printf("Failed to insert word: translate is empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	row := golearn.Row{
		Word:      word,
		Translate: translate,
	}

	err := h.Service.InsertWord(row)
	if err != nil {
		log.Printf("Failed to insert word: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(row)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, string(out))
}
