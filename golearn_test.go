package golearn

import (
	"testing"
)

func TestGetLanguage(t *testing.T) {
	lang, err := GetLanguage("./no_exists_file_with_such_name_ever_in_this_project")
	if err == nil {
		t.Fatalf("error should be not nil for non exists language file")
	}

	lang, err = GetLanguage("./lang.json")
	if err != nil {
		t.Fatalf("failed to get language instance: %v", err)
	}

	codes := []string{"ru", "en"}

	total := 0
	for _, code := range codes {
		phrases, ok := lang[code]
		if !ok {
			t.Fatalf("failed to get phrases for %s language", code)
		}

		phrasesLength := len(phrases)

		if phrasesLength == 0 {
			t.Fatalf("failed to get phrases for %s language", code)
		}

		total += phrasesLength
	}

	if total%len(codes) != 0 {
		t.Error("language phrases count mismatch, some phrase/phrases missed")
	}
}
