package storage

import (
	"fmt"
	"path/filepath"
)

// Paths returns path helpers for a project directory.
type Paths struct {
	ProjectDir string
}

// NewPaths creates a Paths helper for the given project directory.
func NewPaths(projectDir string) *Paths {
	return &Paths{ProjectDir: projectDir}
}

// ProjectYAML returns the path to project.yaml.
func (p *Paths) ProjectYAML() string {
	return filepath.Join(p.ProjectDir, "project.yaml")
}

// WorldYAML returns the path to world.yaml.
func (p *Paths) WorldYAML() string {
	return filepath.Join(p.ProjectDir, "world.yaml")
}

// CharactersDir returns the path to the characters directory.
func (p *Paths) CharactersDir() string {
	return filepath.Join(p.ProjectDir, "characters")
}

// ProtagonistYAML returns the path to the protagonist character file.
func (p *Paths) ProtagonistYAML() string {
	return filepath.Join(p.CharactersDir(), "protagonist.yaml")
}

// TimelineDir returns the path to the timeline directory.
func (p *Paths) TimelineDir() string {
	return filepath.Join(p.ProjectDir, "timeline")
}

// ChapterDir returns the path to a specific chapter's event directory.
func (p *Paths) ChapterDir(chapterNum int) string {
	return filepath.Join(p.TimelineDir(), fmt.Sprintf("chapter-%02d", chapterNum))
}

// EventPath returns the path to a specific event YAML file.
func (p *Paths) EventPath(chapterNum int, eventID string) string {
	return filepath.Join(p.ChapterDir(chapterNum), fmt.Sprintf("%s.yaml", eventID))
}

// GeneratedDir returns the path to generated chapters.
func (p *Paths) GeneratedDir() string {
	return filepath.Join(p.ProjectDir, "generated")
}

// ChapterMarkdown returns the path to a generated chapter markdown file.
func (p *Paths) ChapterMarkdown(chapterNum int) string {
	return filepath.Join(p.GeneratedDir(), fmt.Sprintf("chapter-%02d.md", chapterNum))
}

// StateLock returns the path to the state lock file.
func (p *Paths) StateLock() string {
	return filepath.Join(p.ProjectDir, "state.lock")
}
