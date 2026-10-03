package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSkillDescription(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "no front matter",
			content: "just text\n",
			want:    "",
		},
		{
			name:    "missing description",
			content: "---\nname: s\n---\n",
			want:    "",
		},
		{
			name:    "empty description",
			content: "---\nname: s\ndescription:\n---\n",
			want:    "",
		},
		{
			name:    "plain multi-line description folds",
			content: "---\nname: s\ndescription:\n  first line\n  second line\n---\n",
			want:    "first line second line",
		},
		{
			name:    "indented dashes do not close front matter",
			content: "---\nname: s\ndescription: |\n  first\n  ---\n  second\n---\n",
			want:    "first\n---\nsecond",
		},
		{
			name:    "unterminated front matter",
			content: "---\nname: s\n",
			want:    "",
		},
		{
			name:    "crlf front matter",
			content: "---\r\nname: s\r\ndescription: hello\r\n---\r\n",
			want:    "hello",
		},
		{
			name:    "single-line value",
			content: "---\nname: s\ndescription: Goravel testing skill\n---\n",
			want:    "Goravel testing skill",
		},
		{
			name:    "quoted single-line value",
			content: "---\nname: s\ndescription: \"Goravel testing skill\"\n---\n",
			want:    "Goravel testing skill",
		},
		{
			name:    "quoted multi-line value strips the closing quote",
			content: "---\nname: s\ndescription: \"first\n  second\"\n---\n",
			want:    "first second",
		},
		{
			name:    "inline comment is stripped from an unquoted value",
			content: "---\nname: s\ndescription: hello # comment\n---\n",
			want:    "hello",
		},
		{
			name:    "inline comment is stripped after a quoted value",
			content: "---\nname: s\ndescription: \"hello\" # comment\n---\n",
			want:    "hello",
		},
		{
			name:    "inline value folds continuation lines",
			content: "---\nname: s\ndescription: first line\n  second line\n---\n",
			want:    "first line second line",
		},
		{
			name:    "indented nested description is ignored",
			content: "---\nname: s\nmetadata:\n  description: nested\n---\n",
			want:    "",
		},
		{
			name:    "folded block joins lines",
			content: "---\nname: s\ndescription: >\n  first line\n  second line\n---\n",
			want:    "first line second line",
		},
		{
			name:    "folded block stops at next key",
			content: "---\nname: s\ndescription: >\n  first line\n  second line\nother: x\n---\n",
			want:    "first line second line",
		},
		{
			name:    "literal block keeps newlines",
			content: "---\nname: s\ndescription: |\n  first line\n  second line\n---\n",
			want:    "first line\nsecond line",
		},
		{
			name:    "literal block with blank edges",
			content: "---\nname: s\ndescription: |\n\n  first line\n  second line\n\n---\n",
			want:    "first line\nsecond line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseSkillDescription(tt.content))
		})
	}
}

func TestReadSkillDescriptionStopsAtFrontMatter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	body := "---\nname: s\ndescription: hello\n---\n" + strings.Repeat("x", 2<<20)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))

	got, err := readSkillDescription(path)
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestReadSkillDescriptionPlainTextBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(path, []byte("plain body without front matter\n"), 0o644))

	got, err := readSkillDescription(path)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestReadSkillDescriptionRejectsNonRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.MkdirAll(path, 0o755))

	_, err := readSkillDescription(path)
	assert.ErrorContains(t, err, "is not a regular file")
}
