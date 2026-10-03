package commands

import "errors"

// noSkillsFoundMessage is shared by skill:install and skill:list so both report a
// consistent error when the goravel/agents repo yields no skills.
const noSkillsFoundMessage = "no skills found in goravel/agents"

// errNoSkillsFound is returned when the goravel/agents repo yields no skills.
var errNoSkillsFound = errors.New(noSkillsFoundMessage)
