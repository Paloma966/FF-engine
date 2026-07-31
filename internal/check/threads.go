package check

import (
	"fmt"

	"github.com/chun/fiction_factory/internal/models"
)

const maxStalledEvents = 5 // Events before a hook is flagged as stalled

// CheckThreads detects unresolved hooks that have been abandoned or neglected.
func CheckThreads(world *models.WorldState, events []models.Event) []Issue {
	var issues []Issue

	unresolved := world.UnresolvedThreads()

	// Build a reverse index: hook ID → events that reference it
	hookReferences := make(map[string][]string)
	for _, evt := range events {
		for _, h := range evt.FutureHooks {
			hookReferences[h.ID] = append(hookReferences[h.ID], evt.ID)
		}
	}

	// Find events that resolve hooks and check they actually do
	for _, evt := range events {
		for _, resolvedID := range evt.ResolvesHooks {
			// Check the resolved hook exists
			found := false
			for _, t := range world.Threads {
				if t.ID == resolvedID {
					found = true
					break
				}
			}
			if !found {
				issues = append(issues, Issue{
					Severity: SeverityWarning,
					Category: "threads",
					Message: fmt.Sprintf(
						"Event %s resolves hook '%s' but this hook is not registered in world threads",
						evt.ID, resolvedID,
					),
					RelevantEvents: []string{evt.ID},
				})
			}

			// Check the hook was actually planted before this event
			for _, t := range world.Threads {
				if t.ID == resolvedID && t.PlantedIn != "" {
					plantedEvent := findEventByID(events, t.PlantedIn)
					if plantedEvent != nil && plantedEvent.ChapterNum >= evt.ChapterNum {
						issues = append(issues, Issue{
							Severity: SeverityWarning,
							Category: "threads",
							Message: fmt.Sprintf(
								"Event %s resolves hook '%s' but the hook was planted in %s (same or later chapter)",
								evt.ID, resolvedID, t.PlantedIn,
							),
							RelevantEvents: []string{evt.ID, t.PlantedIn},
						})
					}
				}
			}
		}
	}

	// Check each unresolved hook
	for _, hook := range unresolved {
		// Count events since planting
		eventsSince := countEventsSince(events, hook.PlantedIn)

		// Immediate hooks that are stalled
		if hook.Urgency == "immediate" {
			if eventsSince > maxStalledEvents {
				issues = append(issues, Issue{
					Severity: SeverityWarning,
					Category: "threads",
					Message: fmt.Sprintf(
						"Hook '%s' marked as 'immediate' urgency but unresolved after %d events",
						hook.ID, eventsSince,
					),
					RelevantEvents: []string{hook.PlantedIn},
				})
			} else if eventsSince >= 3 {
				issues = append(issues, Issue{
					Severity: SeverityInfo,
					Category: "threads",
					Message: fmt.Sprintf(
						"Hook '%s' marked as 'immediate' — %d events since planting, consider paying it off soon",
						hook.ID, eventsSince,
					),
					RelevantEvents: []string{hook.PlantedIn},
				})
			}
		}

		// Hooks that were never referenced again after planting
		if refs, ok := hookReferences[hook.ID]; !ok || len(refs) <= 1 {
			// Only one reference (the planting itself) — this hook is gathering dust
			if eventsSince > 10 {
				issues = append(issues, Issue{
					Severity: SeverityWarning,
					Category: "threads",
					Message: fmt.Sprintf(
						"Hook '%s' was planted %d events ago and has never been referenced since",
						hook.ID, eventsSince,
					),
					RelevantEvents: []string{hook.PlantedIn},
				})
			}
		}

		// "Dormant" hooks that are very old
		if hook.Urgency == "dormant" && eventsSince > 20 {
			issues = append(issues, Issue{
				Severity: SeverityInfo,
				Category: "threads",
				Message: fmt.Sprintf(
					"Hook '%s' is dormant and was planted %d events ago — consider resolving or dropping",
					hook.ID, eventsSince,
				),
				RelevantEvents: []string{hook.PlantedIn},
			})
		}
	}

	// Check for hooks that are duplicated
	seenHooks := make(map[string]string)
	for _, evt := range events {
		for _, h := range evt.FutureHooks {
			if prevEvent, ok := seenHooks[h.ID]; ok {
				issues = append(issues, Issue{
					Severity: SeverityError,
					Category: "threads",
					Message: fmt.Sprintf(
						"Duplicate hook ID '%s' — planted in both %s and %s",
						h.ID, prevEvent, evt.ID,
					),
					RelevantEvents: []string{prevEvent, evt.ID},
				})
			}
			seenHooks[h.ID] = evt.ID
		}
	}

	return issues
}

// countEventsSince returns how many events have occurred since the given event ID.
func countEventsSince(events []models.Event, eventID string) int {
	if eventID == "" {
		return 0
	}

	plantedIndex := -1
	for i, e := range events {
		if e.ID == eventID {
			plantedIndex = i
			break
		}
	}
	if plantedIndex < 0 {
		return 0
	}
	return len(events) - plantedIndex - 1
}

// findEventByID finds an event by its ID.
func findEventByID(events []models.Event, id string) *models.Event {
	for i := range events {
		if events[i].ID == id {
			return &events[i]
		}
	}
	return nil
}
