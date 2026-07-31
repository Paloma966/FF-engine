package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chun/fiction_factory/internal/llm"
	"github.com/chun/fiction_factory/internal/models"
	"github.com/chun/fiction_factory/internal/storage"
)

// RunLoop orchestrates one iteration of the story production cycle:
// Director → [User Review] → Protagonist → Chapter → Save.
type RunLoop struct {
	director     *DirectorAgent
	protagonist  *ProtagonistAgent
	chapter      *ChapterGenerator
	loader       *storage.Loader
	saver        *storage.Saver
	llmDirector  llm.LLMClient
	llmProtagonist llm.LLMClient
	llmChapter   llm.LLMClient
}

// LoopConfig configures the RunLoop with potentially different LLM clients per agent.
type LoopConfig struct {
	DirectorLLM    llm.LLMClient
	ProtagonistLLM llm.LLMClient
	ChapterLLM     llm.LLMClient
	POV            string
	Tense          string
}

// NewRunLoop creates a new RunLoop.
func NewRunLoop(loader *storage.Loader, saver *storage.Saver, cfg LoopConfig) *RunLoop {
	return &RunLoop{
		director:       NewDirectorAgent(cfg.DirectorLLM),
		protagonist:    NewProtagonistAgent(cfg.ProtagonistLLM),
		chapter:        NewChapterGenerator(cfg.ChapterLLM),
		loader:         loader,
		saver:          saver,
		llmDirector:    cfg.DirectorLLM,
		llmProtagonist: cfg.ProtagonistLLM,
		llmChapter:     cfg.ChapterLLM,
	}
}

// RunResult contains all outputs from a single run iteration.
type RunResult struct {
	Event    *models.Event
	Analysis *DirectorOutput
	Reaction *models.ProtagonistReaction
	Chapter  string
}

// POV returns the project's point of view setting.
func (rl *RunLoop) pov() string {
	proj, err := rl.loader.LoadProject()
	if err != nil {
		return "third_person_limited"
	}
	return proj.Story.POV
}

// tense returns the project's tense setting.
func (rl *RunLoop) tense() string {
	proj, err := rl.loader.LoadProject()
	if err != nil {
		return "past"
	}
	return proj.Story.Tense
}

// Step1_DirectorProposal runs the Director Agent and returns proposed events.
func (rl *RunLoop) Step1_DirectorProposal(ctx context.Context) (*models.Event, *DirectorOutput, error) {
	// Load all state
	world, err := rl.loader.LoadWorld()
	if err != nil {
		return nil, nil, fmt.Errorf("load world: %w", err)
	}
	protagonist, err := rl.loader.LoadProtagonist()
	if err != nil {
		return nil, nil, fmt.Errorf("load protagonist: %w", err)
	}
	recentEvents, err := rl.loader.LoadRecentEvents(5)
	if err != nil {
		return nil, nil, fmt.Errorf("load recent events: %w", err)
	}

	nextChapter, err := rl.loader.NextChapterNum()
	if err != nil {
		return nil, nil, fmt.Errorf("get next chapter: %w", err)
	}

	// Build state
	state := DirectorState{
		Facts:            world.Facts,
		Threads:          world.UnresolvedThreads(),
		Protagonist:      protagonist,
		RecentMemories:   protagonist.RecentMemories(5),
		RecentEvents:     recentEvents,
		RecentEventCount: len(recentEvents),
		NextChapterNum:   nextChapter,
		CurrentTime:      world.CurrentNarrativeTime,
	}

	return rl.director.ProposeEvent(ctx, state)
}

// Step2_ProtagonistResponse runs the Protagonist Agent on an event.
func (rl *RunLoop) Step2_ProtagonistResponse(ctx context.Context, event *models.Event) (*models.ProtagonistReaction, error) {
	protagonist, err := rl.loader.LoadProtagonist()
	if err != nil {
		return nil, fmt.Errorf("load protagonist: %w", err)
	}

	state := ProtagonistState{
		Name:           protagonist.Name,
		Beliefs:        protagonist.Beliefs,
		Goals:          protagonist.Goals,
		Fears:          protagonist.Fears,
		Values:         protagonist.Values,
		RecentMemories: protagonist.RecentMemories(5),
		Event: ProcessedEvent{
			Title:          event.Title,
			Time:           event.Time,
			Tone:           event.Tone,
			Description:    event.Description,
			FactsChanged:   event.FactsChanged,
			Participants:   event.Participants,
			DirectorIntent: event.DirectorIntent,
		},
	}

	return rl.protagonist.ProcessEvent(ctx, state)
}

// Step3_GenerateChapter generates the chapter prose.
func (rl *RunLoop) Step3_GenerateChapter(ctx context.Context, event *models.Event, reaction *models.ProtagonistReaction) (string, error) {
	input := ChapterInputFromEvent(event, reaction, rl.pov(), rl.tense(), "")

	// Add chapter number header if not present
	chapter, err := rl.chapter.Generate(ctx, input)
	if err != nil {
		return "", err
	}

	// Ensure chapter has a title
	if !strings.HasPrefix(chapter, "#") {
		chapter = fmt.Sprintf("# Chapter %d: %s\n\n%s", event.ChapterNum, event.Title, chapter)
	}

	return chapter, nil
}

// SaveAll persists all outputs from a run iteration.
func (rl *RunLoop) SaveAll(event *models.Event, reaction *models.ProtagonistReaction, chapterText string) error {
	// Set IDs and chapter numbers
	world, err := rl.loader.LoadWorld()
	if err != nil {
		return fmt.Errorf("load world: %w", err)
	}

	eventID, err := rl.loader.NextEventID()
	if err != nil {
		return fmt.Errorf("next event ID: %w", err)
	}

	event.ID = eventID
	event.ChapterNum = world.ChapterCount + 1
	event.Lifecycle = models.EventResolved
	event.ProtagonistResponse = reaction
	event.ChapterText = chapterText

	// Fill in derived FutureHook fields
	for i := range event.FutureHooks {
		event.FutureHooks[i].PlantedIn = eventID
	}

	// Save event
	if err := rl.saver.SaveEvent(event); err != nil {
		return fmt.Errorf("save event: %w", err)
	}

	// Update and save world state
	now := time.Now().Format("2006-01-02 15:04")
	rl.applyFactsToWorld(world, event)
	world.Threads = append(world.Threads, event.FutureHooks...)
	rl.markResolvedHooks(world, event.ResolvesHooks)
	world.ChapterCount = event.ChapterNum
	world.EventCount++
	world.CurrentNarrativeTime = fmt.Sprintf("%s (Event: %s, %s)", event.Time, eventID, now)

	if err := rl.saver.SaveWorld(world); err != nil {
		return fmt.Errorf("save world: %w", err)
	}

	// Update and save protagonist
	protagonist, err := rl.loader.LoadProtagonist()
	if err != nil {
		return fmt.Errorf("load protagonist: %w", err)
	}
	ApplyReaction(protagonist, eventID, reaction, event.Time)
	if err := rl.saver.SaveProtagonist(protagonist); err != nil {
		return fmt.Errorf("save protagonist: %w", err)
	}

	// Save chapter markdown
	if err := rl.saver.SaveChapterMarkdown(event.ChapterNum, chapterText); err != nil {
		return fmt.Errorf("save chapter: %w", err)
	}

	return nil
}

// applyFactsToWorld updates world facts based on event's FactsChanged.
func (rl *RunLoop) applyFactsToWorld(world *models.WorldState, event *models.Event) {
	for _, fc := range event.FactsChanged {
		// Remove old fact
		world.Facts = removeString(world.Facts, fc.Before)
		// Add new fact
		world.Facts = append(world.Facts, fc.After)
	}
}

// markResolvedHooks marks hooks as resolved in the world state.
func (rl *RunLoop) markResolvedHooks(world *models.WorldState, resolvedIDs []string) {
	for _, id := range resolvedIDs {
		for i := range world.Threads {
			if world.Threads[i].ID == id {
				world.Threads[i].ResolvedIn = "resolved"
			}
		}
	}
}

// RunFull executes the complete cycle and saves everything.
func (rl *RunLoop) RunFull(ctx context.Context) (*RunResult, error) {
	// Step 1: Director
	fmt.Println("\n🎬 Director is analyzing the story...")
	event, analysis, err := rl.Step1_DirectorProposal(ctx)
	if err != nil {
		return nil, fmt.Errorf("director step: %w", err)
	}

	// Step 2: Protagonist
	fmt.Println("🎭 Protagonist is processing the event...")
	reaction, err := rl.Step2_ProtagonistResponse(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("protagonist step: %w", err)
	}

	// Step 3: Chapter
	fmt.Println("📝 Generating chapter prose...")
	chapter, err := rl.Step3_GenerateChapter(ctx, event, reaction)
	if err != nil {
		return nil, fmt.Errorf("chapter step: %w", err)
	}

	// Save
	fmt.Println("💾 Saving...")
	if err := rl.SaveAll(event, reaction, chapter); err != nil {
		return nil, fmt.Errorf("save step: %w", err)
	}

	return &RunResult{
		Event:    event,
		Analysis: analysis,
		Reaction: reaction,
		Chapter:  chapter,
	}, nil
}
