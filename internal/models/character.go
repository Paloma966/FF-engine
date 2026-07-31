package models

// Memory is a single remembered experience stored in the Character.
type Memory struct {
	ID        string   `yaml:"id"`
	EventID   string   `yaml:"event_id"`
	Content   string   `yaml:"content"`
	Intensity int      `yaml:"intensity"` // 1-10
	Category  string   `yaml:"category"`  // trauma | triumph | revelation | connection | loss
	Triggers  []string `yaml:"triggers"`
	Timestamp string   `yaml:"timestamp"`
}

// Character represents a character in the story (MVP: protagonist only).
type Character struct {
	Name             string   `yaml:"name"`
	Beliefs          []string `yaml:"beliefs"`
	Goals            []string `yaml:"goals"`
	Fears            []string `yaml:"fears"`
	Values           []string `yaml:"values"`
	Memories         []Memory `yaml:"memories"`
	AbandonedBeliefs []string `yaml:"abandoned_beliefs"`
}

// CurrentState returns a human-readable summary of the character's
// current psychological state.
func (c *Character) CurrentState() []string {
	state := make([]string, 0, len(c.Beliefs)+len(c.Goals)+len(c.Fears))
	for _, b := range c.Beliefs {
		state = append(state, "believes: "+b)
	}
	for _, g := range c.Goals {
		state = append(state, "wants: "+g)
	}
	for _, f := range c.Fears {
		state = append(state, "fears: "+f)
	}
	return state
}

// RecentMemories returns the most recent N memories sorted by recency.
// For MVP, we simply return the last N entries (memories are appended in order).
func (c *Character) RecentMemories(n int) []Memory {
	if n <= 0 || len(c.Memories) == 0 {
		return nil
	}
	start := len(c.Memories) - n
	if start < 0 {
		start = 0
	}
	return c.Memories[start:]
}
