package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateForLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		expected string
	}{
		{"短いテキスト", "hello", 50, "hello"},
		{"ちょうど上限", "12345", 5, "12345"},
		{"上限超え", "123456", 5, "12345..."},
		{"日本語テキスト", "あいうえおかきくけこ", 5, "あいうえお..."},
		{"空文字列", "", 50, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateForLog(tt.input, tt.maxRunes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
