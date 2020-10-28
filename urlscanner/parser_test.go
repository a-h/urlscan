package urlscanner

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple URL",
			input:    "https://example.com",
			expected: []string{"https://example.com"},
		},
		{
			name:     "url prefixed with space",
			input:    "   https://example.com",
			expected: []string{"https://example.com"},
		},
		{
			name:     "url surrounded by space",
			input:    "   https://example.com  ",
			expected: []string{"https://example.com"},
		},
		{
			name:     "urls on new lines",
			input:    "   https://example1.com  \n https://example2.com  ",
			expected: []string{"https://example1.com", "https://example2.com"},
		},
		{
			name:     "url-like URLs",
			input:    "   example1.com  \n https://example2.com  ",
			expected: []string{"example1.com", "https://example2.com"},
		},
		{
			name:  "a single word",
			input: "word",
		},
		{
			name:  "just words, no URLs",
			input: "a word is not a URL. even with a fullstop.",
		},
		{
			name:     "a scheme to a local network host",
			input:    "https://laptop",
			expected: []string{"https://laptop"},
		},
		{
			name:  "a mistyped sentence",
			input: "I mistyped.the sentence.",
		},
		{
			name:  "lots of words",
			input: "shut.the.front.door",
		},
		{
			name:     "url surrounded by <>",
			input:    "<https://laptop>",
			expected: []string{"https://laptop"},
		},
		{
			name:     "url in a sentence",
			input:    "Head over to https://sentence/test and see what you think",
			expected: []string{"https://sentence/test"},
		},
		{
			name:     "multiple urls in sentences",
			input:    "Head over to https://sentence2/test and see what you think. Also, cast your eye over example.com and report back.",
			expected: []string{"https://sentence2/test", "example.com"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Scan(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("input\n%s\n\nexpected\n%s\n\nactual\n%s\n", tt.input,
					output(tt.expected), output(actual))
			}
		})
	}
}

func output(values []string) string {
	op := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		op[i] = fmt.Sprintf("%q", values[i])
	}
	return strings.Join(op, ", ")
}
