package golearn

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestGetLanguage(t *testing.T) {
	t.Run("Check language file content test", func(t *testing.T) {
		codes := []string{"ru", "en"}

		total := 0
		for _, code := range codes {
			langFilename := fmt.Sprintf("./lang.%s.json", code)
			langContent, err := ioutil.ReadFile(langFilename)
			if err != nil {
				t.Fatalf("failed to read language file: %s, %v", langFilename, err)
			}
			lang, err := GetLanguage(langContent)
			if err != nil {
				t.Fatalf("failed to get language instance: %v", err)
			}

			phrasesLength := len(lang)

			if phrasesLength == 0 {
				t.Fatalf("failed to get phrases for %s language", code)
			}

			total += phrasesLength
		}

		if total%len(codes) != 0 {
			t.Error("language phrases count mismatch, some phrase/phrases missed")
		}
	})
}
