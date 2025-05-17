package envfile

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/support/file"
)

func ReplaceValues(filepath string, replacements map[string]string) error {
	src, err := file.GetContent(filepath)
	if err != nil {
		return err
	}

	contest := strings.Split(src, "\n")
	position := make(map[string]int)

	for i, line := range contest {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, "=") {
			key := strings.Split(line, "=")[0]
			position[key] = i
		}
	}

	for key, value := range replacements {
		if strings.Contains(value, " ") {
			value = fmt.Sprintf(`"%s"`, value)
		}

		if pos, ok := position[key]; ok {
			contest[pos] = fmt.Sprintf("%s=%s", key, value)
			continue
		}

		contest = append(contest, fmt.Sprintf("%s=%s", key, value))
	}

	return file.PutContent(filepath, strings.Join(contest, "\n"))
}
