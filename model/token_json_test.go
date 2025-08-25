package model

import (
	"encoding/json"
	"testing"

	"github.com/songquanpeng/one-api/common/config"
)

func TestTokenMarshalJSON_DefaultPrefix(t *testing.T) {
	// backup and restore
	old := config.TokenKeyPrefix
	config.TokenKeyPrefix = "sk-"
	defer func() { config.TokenKeyPrefix = old }()

	tok := Token{Id: 1, UserId: 2, Key: "abcdef"}
	b, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if got := string(b); !containsJSONPair(got, `"key":"sk-abcdef"`) {
		t.Fatalf("expected key with sk- prefix, got: %s", got)
	}
}

func TestTokenMarshalJSON_CustomPrefix(t *testing.T) {
	old := config.TokenKeyPrefix
	config.TokenKeyPrefix = "custom-"
	defer func() { config.TokenKeyPrefix = old }()

	tok := Token{Id: 1, UserId: 2, Key: "abcdef"}
	b, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if got := string(b); !containsJSONPair(got, `"key":"custom-abcdef"`) {
		t.Fatalf("expected key with custom- prefix, got: %s", got)
	}
}

func TestTokenMarshalJSON_StripsLegacyPrefix(t *testing.T) {
	old := config.TokenKeyPrefix
	config.TokenKeyPrefix = "sk-"
	defer func() { config.TokenKeyPrefix = old }()

	tok := Token{Id: 1, UserId: 2, Key: "sk-abcdef"}
	b, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if got := string(b); !containsJSONPair(got, `"key":"sk-abcdef"`) {
		t.Fatalf("expected single sk- prefix, got: %s", got)
	}
}

// containsJSONPair is a tiny helper to avoid pulling extra deps
func containsJSONPair(s, pair string) bool {
	return len(s) >= len(pair) && (stringContains(s, pair))
}

func stringContains(s, sub string) bool {
	return (len(sub) == 0) || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	// simple search to avoid importing strings for a single use in tests
outer:
	for i := 0; i+len(sub) <= len(s); i++ {
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
