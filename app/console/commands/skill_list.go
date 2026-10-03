package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

type skillDetail struct {
	Name        string
	Description string
}

// fetchSkills clones the agents repo and returns its skills, optionally with
// descriptions. It errors when the repo yields no skills.
func fetchSkills(detail bool) ([]skillDetail, error) {
	var skills []skillDetail

	err := withAgentsRepo(func(skillsPath string) error {
		var err error
		skills, err = listSkillDetails(skillsPath, detail)
		if err != nil {
			return err
		}
		if len(skills) == 0 {
			return errNoSkillsFound
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return skills, nil
}

// listSkills returns the names of the skill directories under skillsPath. A
// missing skills folder is treated as an empty result so callers can surface
// their own "no skills found" message. Symlinked entries are rejected with the
// same "not a directory" error used for explicitly named skills, and other
// non-directory entries are skipped.
func listSkills(skillsPath string) ([]string, error) {
	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read skills: %w", err)
	}

	skills := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		// Hidden entries (e.g. dot-prefixed staging/backup leftovers) are not skills.
		if name[0] == '.' {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("skill %q is not a directory", name)
		}
		if entry.IsDir() {
			skills = append(skills, name)
		}
	}

	return skills, nil
}

// listSkillDetails returns the repo's skills, optionally with each skill's
// description. A skill whose SKILL.md is missing or unreadable is kept with an
// empty description instead of aborting the whole listing.
func listSkillDetails(skillsPath string, detail bool) ([]skillDetail, error) {
	skillNames, err := listSkills(skillsPath)
	if err != nil {
		return nil, err
	}

	skills := make([]skillDetail, 0, len(skillNames))
	for _, skillName := range skillNames {
		skill := skillDetail{Name: skillName}
		if detail {
			if description, err := readSkillDescription(filepath.Join(skillsPath, skillName, "SKILL.md")); err == nil {
				skill.Description = description
			}
		}

		skills = append(skills, skill)
	}

	return skills, nil
}
