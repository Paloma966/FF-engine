package prompts

import (
	_ "embed"
)

// Director prompts
//
//go:embed director_system.txt
var DirectorSystem string

//go:embed director_task.txt
var DirectorTaskTemplate string

// Protagonist prompts
//
//go:embed protagonist_system.txt
var ProtagonistSystem string

//go:embed protagonist_task.txt
var ProtagonistTaskTemplate string

// Chapter prompts
//
//go:embed chapter_system.txt
var ChapterSystem string

//go:embed chapter_task.txt
var ChapterTaskTemplate string
