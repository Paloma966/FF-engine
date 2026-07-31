package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chun/fiction_factory/internal/engine"
	"github.com/chun/fiction_factory/internal/llm"
	"github.com/chun/fiction_factory/internal/models"
	"github.com/chun/fiction_factory/internal/storage"
	"github.com/spf13/cobra"
)

var (
	runAuto      bool
	runSkipReview bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the story production loop (Director → Protagonist → Chapter)",
	Long: `Execute one iteration of the core story production loop:

  1. Director Agent analyzes the world state and proposes events
  2. You review the Director's analysis and approve the event
  3. Protagonist Agent processes the event psychologically
  4. Chapter Generator produces the final prose
  5. All outputs are saved to the project directory

Each run produces one chapter. The generated content is saved under
timeline/ (event data) and generated/ (chapter markdown).

Examples:
  ff run                  # Interactive: review and approve each step
  ff run --skip-review    # Skip review, auto-approve Director's choice
  ff run --auto           # Non-interactive: generate without prompts`,
	RunE: runRun,
}

func init() {
	runCmd.Flags().BoolVar(&runAuto, "auto", false, "Run non-interactively (no user prompts)")
	runCmd.Flags().BoolVar(&runSkipReview, "skip-review", false, "Skip review of Director analysis and approval step")
}

func runRun(cmd *cobra.Command, args []string) error {
	projectDir := GetProjectDir()
	paths := storage.NewPaths(projectDir)

	// Check project exists
	if _, err := os.Stat(paths.ProjectYAML()); os.IsNotExist(err) {
		return fmt.Errorf("no project found at %s — run 'ff init' first", projectDir)
	}

	loader := storage.NewLoader(paths)
	saver := storage.NewSaver(paths)

	// Load project config
	proj, err := loader.LoadProject()
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}

	// Resolve API key from env if needed
	apiKey := resolveAPIKeyFromEnv(proj.LLM.APIKey)
	proj.LLM.APIKey = apiKey

	// Create LLM clients
	llmCfg := llm.LLMConfig{
		Provider:    proj.LLM.Provider,
		Model:       proj.LLM.Model,
		APIKey:      apiKey,
		BaseURL:     proj.LLM.BaseURL,
		Temperature: proj.LLM.Temperature,
		MaxTokens:   proj.LLM.MaxTokens,
	}

	directorLLM, err := llm.NewFromConfig(llmCfg)
	if err != nil {
		return fmt.Errorf("create director LLM: %w", err)
	}
	protagonistLLM, err := llm.NewFromConfig(llmCfg)
	if err != nil {
		return fmt.Errorf("create protagonist LLM: %w", err)
	}
	chapterLLM, err := llm.NewFromConfig(llmCfg)
	if err != nil {
		return fmt.Errorf("create chapter LLM: %w", err)
	}

	loop := engine.NewRunLoop(loader, saver, engine.LoopConfig{
		DirectorLLM:    directorLLM,
		ProtagonistLLM: protagonistLLM,
		ChapterLLM:     chapterLLM,
		POV:            proj.Story.POV,
		Tense:          proj.Story.Tense,
	})

	ctx := context.Background()

	// --- Step 1: Director Analysis ---
	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("🎬 STEP 1: DIRECTOR ANALYSIS")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()

	event, analysis, err := loop.Step1_DirectorProposal(ctx)
	if err != nil {
		return fmt.Errorf("director: %w", err)
	}

	// Show full analysis
	analysis.PrintAnalysis()

	// Interactive review
	if !runAuto && !runSkipReview {
		if !confirm("\n▶ Accept this event and continue?") {
			fmt.Println("\n⚠️  Event rejected. No changes saved.")
			return nil
		}
	}

	// --- Step 2: Protagonist Response ---
	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("🎭 STEP 2: PROTAGONIST RESPONSE")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()

	reaction, err := loop.Step2_ProtagonistResponse(ctx, event)
	if err != nil {
		return fmt.Errorf("protagonist: %w", err)
	}

	engine.PrintReaction(reaction, event.Title)

	if !runAuto && !runSkipReview {
		if !confirm("\n▶ Accept this character response and continue?") {
			fmt.Println("\n⚠️  Response rejected. No changes saved.")
			return nil
		}
	}

	// --- Step 3: Chapter Generation ---
	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("📝 STEP 3: CHAPTER GENERATION")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()

	chapter, err := loop.Step3_GenerateChapter(ctx, event, reaction)
	if err != nil {
		return fmt.Errorf("chapter: %w", err)
	}

	// Preview first few lines
	preview := chapter
	if len(preview) > 500 {
		preview = preview[:500] + "\n...\n[truncated]"
	}
	fmt.Println(preview)
	fmt.Println()

	if !runAuto && !runSkipReview {
		if !confirm("\n▶ Save this chapter?") {
			fmt.Println("\n⚠️  Chapter discarded. No changes saved.")
			return nil
		}
	}

	// --- Step 4: Save ---
	fmt.Println("\n💾 Saving...")
	if err := loop.SaveAll(event, reaction, chapter); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("✅ Chapter %d complete!\n", event.ChapterNum)
	fmt.Printf("   Event: %s (%s)\n", event.ID, event.Title)
	fmt.Printf("   Saved: timeline/chapter-%02d/%s.yaml\n", event.ChapterNum, event.ID)
	fmt.Printf("   Saved: generated/chapter-%02d.md\n", event.ChapterNum)
	fmt.Println(strings.Repeat("═", 60))

	return nil
}

// confirm asks the user a yes/no question and returns true for yes.
func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [Y/n]: ", prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

// resolveAPIKeyFromEnv replaces env var references with actual values.
func resolveAPIKeyFromEnv(key string) string {
	if strings.HasPrefix(key, "${") && strings.HasSuffix(key, "}") {
		envName := key[2 : len(key)-1]
		return os.Getenv(envName)
	}
	return key
}

// --- Test helpers for the run command ---

// createTestProject scaffolds a minimal project for testing.
func createTestProject(dir, name, provider string) error {
	paths := storage.NewPaths(dir)
	saver := storage.NewSaver(paths)
	if err := saver.EnsureDirs(); err != nil {
		return err
	}

	proj := &models.Project{
		Name:      name,
		CreatedAt: "2026-07-30T00:00:00Z",
		Story: models.StoryConfig{
			Title:   name,
			Genre:   "fantasy",
			POV:     "third_person_limited",
			Tense:   "past",
		},
		Protagonist: models.ProtagonistConfig{
			Name: "Lin En",
		},
		LLM: models.LLMConfig{
			Provider:    provider,
			Model:       "test-model",
			APIKey:      "test-key",
			Temperature: 0.8,
			MaxTokens:   4096,
		},
	}
	if err := saver.SaveProject(proj); err != nil {
		return err
	}

	world := &models.WorldState{
		Facts:                []string{},
		Threads:              nil,
		ChapterCount:         0,
		EventCount:           0,
		CurrentNarrativeTime: "Day 0, Prologue",
	}
	if err := saver.SaveWorld(world); err != nil {
		return err
	}

	char := &models.Character{
		Name: "Lin En",
	}
	if err := saver.SaveProtagonist(char); err != nil {
		return err
	}

	return nil
}
