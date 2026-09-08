// Package syncdue decides which participant sources are ready to run.
//
// Separate from the worker that runs them, because "is this due" is arithmetic
// over a schedule and a last run, and arithmetic that decides whether a thing
// happens should be testable without a database, a clock or a network.
package syncdue

import (
	"time"

	"github.com/gsoultan/metis/server/domains/entities"
)

// RetryAfter is how long a failed sync waits before being tried again.
//
// Independent of the source's own schedule, and deliberately shorter than most
// of them: a directory that was briefly unreachable should be picked up soon,
// and one that is genuinely broken should not be hammered every tick. A source
// on a daily schedule that fails at midnight would otherwise stay broken until
// the following midnight.
const RetryAfter = 15 * time.Minute

// Due reports whether a source should run now.
//
// A source that has never run is due immediately. Somebody who has just
// configured a directory expects it to fill in, not to sit empty until the
// first interval elapses — and waiting would make a mistyped query look like a
// schedule that has not fired yet.
func Due(source entities.ParticipantSource, now time.Time) bool {
	if !source.Enabled || source.Schedule == "" {
		return false
	}
	if source.LastRun == nil {
		return true
	}
	if !source.LastRun.OK {
		return !source.LastRun.At.Add(RetryAfter).After(now)
	}
	every, ok := Interval(source.Schedule)
	if !ok {
		// An unreadable schedule is not a reason to run constantly. It is
		// refused when the source is saved, so reaching here means the stored
		// value was edited past that check.
		return false
	}
	return !source.LastRun.At.Add(every).After(now)
}

// Interval reads how often a schedule repeats.
//
// The schedule is an ISO 8601 repeating interval — "R/PT1H" — which is the
// vocabulary BPMN timers already use in this system, so somebody who has set a
// timer on a process already knows how to write one.
func Interval(schedule string) (time.Duration, bool) {
	parsed, err := entities.ParseTimerSchedule(schedule, time.Time{})
	if err != nil || parsed.Every <= 0 {
		return 0, false
	}
	return parsed.Every, true
}

// Ready filters a list to the sources that should run now.
func Ready(sources []entities.ParticipantSource, now time.Time) []entities.ParticipantSource {
	var due []entities.ParticipantSource
	for _, source := range sources {
		if Due(source, now) {
			due = append(due, source)
		}
	}
	return due
}
