package syncdue_test

import (
	"testing"
	"time"

	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/domains/logic/syncdue"
)

var now = time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

func source(over func(*entities.ParticipantSource)) entities.ParticipantSource {
	s := entities.ParticipantSource{Enabled: true, Schedule: "R/PT1H"}
	if over != nil {
		over(&s)
	}
	return s
}

// A source nobody has run yet runs immediately: somebody who has just
// configured a directory expects it to fill in, not to sit empty until the
// first interval elapses.
func TestANewSourceRunsImmediately(t *testing.T) {
	if !syncdue.Due(source(nil), now) {
		t.Fatal("a source that has never run should be due")
	}
}

func TestASourceRunsOnceItsIntervalHasElapsed(t *testing.T) {
	justRan := source(func(s *entities.ParticipantSource) {
		s.LastRun = &entities.SourceRun{At: now.Add(-30 * time.Minute), OK: true}
	})
	if syncdue.Due(justRan, now) {
		t.Fatal("half an hour into an hourly schedule is not due")
	}

	overdue := source(func(s *entities.ParticipantSource) {
		s.LastRun = &entities.SourceRun{At: now.Add(-61 * time.Minute), OK: true}
	})
	if !syncdue.Due(overdue, now) {
		t.Fatal("an hour past an hourly schedule is due")
	}
}

// A failure retries sooner than the schedule, and not immediately.
//
// A source on a daily schedule that fails at midnight would otherwise stay
// broken until the following midnight; retrying every tick would hammer a
// directory that is genuinely down.
func TestAFailedSyncRetriesSoonerThanItsSchedule(t *testing.T) {
	daily := func(ago time.Duration) entities.ParticipantSource {
		return source(func(s *entities.ParticipantSource) {
			s.Schedule = "R/P1D"
			s.LastRun = &entities.SourceRun{At: now.Add(-ago), OK: false}
		})
	}
	if syncdue.Due(daily(time.Minute), now) {
		t.Fatal("a failure a minute ago should not retry yet")
	}
	if !syncdue.Due(daily(syncdue.RetryAfter+time.Minute), now) {
		t.Fatal("a failure past the retry window should be tried again, long before the next day")
	}
}

func TestNothingRunsWithoutAScheduleOrWhileDisabled(t *testing.T) {
	manual := source(func(s *entities.ParticipantSource) { s.Schedule = "" })
	if syncdue.Due(manual, now) {
		t.Error("a source with no schedule runs only when somebody asks")
	}
	off := source(func(s *entities.ParticipantSource) { s.Enabled = false })
	if syncdue.Due(off, now) {
		t.Error("a disabled source does not run")
	}
}

// A schedule that will not parse does not mean "run constantly".
func TestAnUnparseableScheduleDoesNotRun(t *testing.T) {
	broken := source(func(s *entities.ParticipantSource) {
		s.Schedule = "every so often"
		s.LastRun = &entities.SourceRun{At: now.Add(-100 * time.Hour), OK: true}
	})
	if syncdue.Due(broken, now) {
		t.Fatal("an unreadable schedule must not run on every tick")
	}
}

func TestReadyFiltersToWhatShouldRun(t *testing.T) {
	sources := []entities.ParticipantSource{
		source(func(s *entities.ParticipantSource) { s.Name = "new" }),
		source(func(s *entities.ParticipantSource) {
			s.Name = "recent"
			s.LastRun = &entities.SourceRun{At: now.Add(-time.Minute), OK: true}
		}),
		source(func(s *entities.ParticipantSource) { s.Name = "off"; s.Enabled = false }),
	}
	due := syncdue.Ready(sources, now)
	if len(due) != 1 || due[0].Name != "new" {
		t.Fatalf("only the new source is due, got %v", due)
	}
}

func TestIntervalReadsTheBPMNVocabulary(t *testing.T) {
	for _, tc := range []struct {
		schedule string
		want     time.Duration
	}{
		{"R/PT1H", time.Hour},
		{"R/PT15M", 15 * time.Minute},
		{"R/P1D", 24 * time.Hour},
	} {
		got, ok := syncdue.Interval(tc.schedule)
		if !ok || got != tc.want {
			t.Errorf("%s should be %v, got %v (ok=%v)", tc.schedule, tc.want, got, ok)
		}
	}
	if _, ok := syncdue.Interval("nonsense"); ok {
		t.Error("an unreadable schedule has no interval")
	}
}
