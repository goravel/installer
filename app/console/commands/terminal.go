package commands

import (
	"strings"
	"unicode"
)

// sanitizeForTerminal removes ASCII and Unicode control/format characters
// (including C0, DEL, C1, BOM, zero-width, and format classes such as bidi
// overrides) from s so repo-provided or computed values cannot emit terminal
// escape sequences or spoof output. When keepFormatting is true, tab and newline
// are preserved (for values displayed across lines); otherwise all control
// characters, including newlines, are stripped.
func sanitizeForTerminal(s string, keepFormatting bool) string {
	return strings.Map(func(r rune) rune {
		if keepFormatting && (r == '\n' || r == '\t') {
			return r
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}

		return r
	}, s)
}
