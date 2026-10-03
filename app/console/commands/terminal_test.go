package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeForTerminal(t *testing.T) {
	tests := []struct {
		name           string
		in             string
		keepFormatting bool
		want           string
	}{
		{name: "plain text unchanged", in: "goravel-testing", want: "goravel-testing"},
		{name: "formatting keeps tab and newline", in: "a\tb\nc", keepFormatting: true, want: "a\tb\nc"},
		{name: "name strips newlines", in: "a\nb\tc", want: "abc"},
		{name: "escape stripped", in: "a\x1b[31mb", want: "a[31mb"},
		{name: "carriage return stripped", in: "a\rb", want: "ab"},
		{name: "del stripped", in: "a\x7fb", want: "ab"},
		{name: "c1 control stripped", in: "a\u009bb", want: "ab"},
		{name: "right-to-left override stripped", in: "a\u202eb", want: "ab"},
		{name: "zero-width space stripped", in: "a\u200bb", want: "ab"},
		{name: "byte order mark stripped", in: "\ufeffabc", want: "abc"},
		{name: "formatting keeps text around bidi override", in: "a\u202eb", keepFormatting: true, want: "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeForTerminal(tt.in, tt.keepFormatting))
		})
	}
}
