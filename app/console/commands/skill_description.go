package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// maxSkillDescriptionBytes caps how much of a SKILL.md is read when parsing the
// description for skill:list --detail.
const maxSkillDescriptionBytes = 1 << 20

// readSkillDescription returns the parsed front-matter description of the
// SKILL.md at path. It rejects anything that is not a regular file and reads at
// most maxSkillDescriptionBytes, stopping at the closing front-matter delimiter
// so the (potentially large) body is never read.
func readSkillDescription(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("skill description %q is not a regular file", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	frontMatter, err := readFrontMatter(file)
	if err != nil {
		return "", err
	}

	return parseSkillDescription(frontMatter), nil
}

// readFrontMatter reads at most maxSkillDescriptionBytes from r and returns the
// front-matter block (from the opening delimiter through the closing "---"), or
// an empty string when the content does not start with front matter. Only a
// column-zero "---" opens or closes the block. Scanner read errors are returned,
// including bufio.ErrTooLong when a front-matter line exceeds
// maxSkillDescriptionBytes (oversized front matter is surfaced, not ignored).
func readFrontMatter(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(io.LimitReader(r, maxSkillDescriptionBytes))
	scanner.Buffer(make([]byte, 0, 64*1024), maxSkillDescriptionBytes)

	var b strings.Builder
	started := false
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if !started {
			if line != "---" {
				return "", nil
			}
			started = true
			b.WriteString(line)
			b.WriteByte('\n')

			continue
		}

		b.WriteString(line)
		b.WriteByte('\n')
		if line == "---" {
			return b.String(), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

// parseSkillDescription extracts the front-matter "description" value from a
// SKILL.md. It returns an empty string when there is no front matter, the front
// matter is not closed, or no description is present. Only a column-zero
// "description:" key is recognized. An inline value (with any indented
// continuation folded in) has quotes stripped; ">" and plain multi-line scalars
// are space-folded; "|" literal blocks preserve newlines.
func parseSkillDescription(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return ""
	}

	var block []string
	inline := ""
	quoted := false
	collect := false
	fold := false
	closed := false

	for _, raw := range lines[1:] {
		line := strings.TrimRight(raw, "\r")
		if line == "---" {
			closed = true
			break
		}

		if collect {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || strings.TrimSpace(line) == "" {
				block = append(block, line)
				continue
			}

			collect = false
		}

		if !strings.HasPrefix(line, "description:") {
			continue
		}

		value := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		if value == "" || strings.HasPrefix(value, ">") || strings.HasPrefix(value, "|") {
			collect = true
			fold = !strings.HasPrefix(value, "|")
			continue
		}

		inline, quoted = parseInlineScalar(value)
		collect = true
		fold = true
	}

	if !closed {
		return ""
	}
	if inline != "" {
		folded := strings.Join(strings.Fields(strings.Join(append([]string{inline}, trimBlankEdges(block)...), " ")), " ")
		if quoted {
			// Drop a closing quote that landed on a continuation line.
			folded = strings.Trim(folded, `"'`)
		}

		return folded
	}

	return formatDescription(block, fold)
}

// parseInlineScalar extracts a YAML inline scalar: a quoted value keeps any '#'
// inside the quotes and drops a trailing comment; an unquoted value drops a
// trailing " # ..." comment. It reports whether the value was quoted so a
// multi-line quoted scalar can have its closing quote stripped after folding.
func parseInlineScalar(value string) (string, bool) {
	if value[0] == '"' || value[0] == '\'' {
		quote := value[0]
		if end := strings.IndexByte(value[1:], quote); end >= 0 {
			return value[1 : 1+end], true
		}

		// Unterminated on this line: the scalar continues on indented lines.
		return strings.TrimLeft(value, `"'`), true
	}

	for i := 1; i < len(value); i++ {
		if value[i] == '#' && (value[i-1] == ' ' || value[i-1] == '\t') {
			return strings.TrimSpace(value[:i]), false
		}
	}

	return value, false
}

// formatDescription folds ">" (and plain multi-line) scalars to a single
// space-separated line, and preserves newlines for "|" literal blocks after
// stripping their common indentation.
func formatDescription(lines []string, fold bool) string {
	lines = trimBlankEdges(lines)

	if fold {
		return strings.Join(strings.Fields(strings.Join(lines, " ")), " ")
	}

	indent := commonIndent(lines)
	unindented := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			unindented = append(unindented, "")
			continue
		}

		unindented = append(unindented, line[indent:])
	}

	return strings.Join(unindented, "\n")
}

// trimBlankEdges removes leading and trailing blank (whitespace-only) lines.
func trimBlankEdges(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// commonIndent returns the length of the shortest leading whitespace run among
// non-blank lines.
func commonIndent(lines []string) int {
	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		width := len(line) - len(strings.TrimLeft(line, " \t"))
		if indent == -1 || width < indent {
			indent = width
		}
	}

	if indent < 0 {
		return 0
	}

	return indent
}
