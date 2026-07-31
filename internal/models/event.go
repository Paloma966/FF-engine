package models

// EventLifecycle tracks the stage of an event in the production pipeline.
type EventLifecycle string

const (
	EventProposed  EventLifecycle = "proposed"
	EventPlanned   EventLifecycle = "planned"
	EventExecuting EventLifecycle = "executing"
	EventResolved  EventLifecycle = "resolved"
	EventAbandoned EventLifecycle = "abandoned"
)

// FactChange captures a before/after pair for an objective world truth.
type FactChange struct {
	Before string `yaml:"before"`
	After  string `yaml:"after"`
}

// BeliefChange captures a before/after pair for a character's subjective internal state.
type BeliefChange struct {
	Character string `yaml:"character"`
	Before    string `yaml:"before"`
	After     string `yaml:"after"`
}

// FutureHook is a narrative thread planted now to be paid off later.
type FutureHook struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	HintsAt     string `yaml:"hints_at"`
	Urgency     string `yaml:"urgency"` // "immediate" | "soon" | "eventual" | "dormant"
	PlantedIn   string `yaml:"planted_in"`
	ResolvedIn  string `yaml:"resolved_in,omitempty"`
}

// EventRef is a lightweight pointer to another event for causal graph edges.
type EventRef struct {
	EventID string `yaml:"event_id"`
	Title   string `yaml:"title"`
}

// Event is the fundamental unit of story progression.
type Event struct {
	ID           string         `yaml:"id"`
	Title        string         `yaml:"title"`
	Lifecycle    EventLifecycle `yaml:"lifecycle"`
	ChapterNum   int            `yaml:"chapter_num"`
	Time         string         `yaml:"time"`
	Participants []string       `yaml:"participants"`

	// Core changes (the three pillars)
	FactsChanged  []FactChange   `yaml:"facts_changed"`
	BeliefChanges []BeliefChange `yaml:"belief_changes"`
	FutureHooks   []FutureHook   `yaml:"future_hooks"`

	// Narrative
	Description string `yaml:"description"`

	// Causal graph
	CausedBy      []EventRef `yaml:"caused_by,omitempty"`
	LeadsTo       []EventRef `yaml:"leads_to,omitempty"`
	ResolvesHooks []string   `yaml:"resolves_hooks,omitempty"`

	// Director metadata
	DirectorIntent string `yaml:"director_intent"`
	TensionLevel   int    `yaml:"tension_level"`
	Tone           string `yaml:"tone"`

	// Filled after protagonist processes
	ProtagonistResponse *ProtagonistReaction `yaml:"protagonist_response,omitempty"`

	// Generated output
	ChapterText string `yaml:"chapter_text,omitempty"`
}

// ProtagonistReaction captures the protagonist's complete psychological
// processing of an event.
type ProtagonistReaction struct {
	EmotionalResponse  string   `yaml:"emotional_response"`
	Decision           string   `yaml:"decision"`
	NewBeliefs         []string `yaml:"new_beliefs"`
	ModifiedBeliefs    []string `yaml:"modified_beliefs"`
	AbandonedBeliefs   []string `yaml:"abandoned_beliefs"`
	StrengthenedValues []string `yaml:"strengthened_values"`
	ChallengedValues   []string `yaml:"challenged_values"`
	MemoryFormed       string   `yaml:"memory_formed"`
	InternalMonologue  string   `yaml:"internal_monologue"`
}
