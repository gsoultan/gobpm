package impl

import (
	"testing"
	"time"
)

func TestWorkerSizingDefaults(t *testing.T) {
	t.Setenv(envJobWorkers, "")
	t.Setenv(envJobPollInterval, "")
	t.Setenv(envJobLease, "")

	got := ResolveJobWorkerSettings()
	if got.Workers != defaultJobWorkers {
		t.Fatalf("workers %d, want %d", got.Workers, defaultJobWorkers)
	}
	if got.PollInterval != defaultJobPollInterval {
		t.Fatalf("poll %s", got.PollInterval)
	}
	if got.Lease != defaultJobLease {
		t.Fatalf("lease %s", got.Lease)
	}
}

func TestWorkerSizingIsConfigurable(t *testing.T) {
	// The point of the change: throughput used to be a compile-time constant of
	// five workers every two seconds, about 2.5 jobs a second per replica, with
	// no way to raise it short of a rebuild.
	t.Setenv(envJobWorkers, "40")
	t.Setenv(envJobPollInterval, "500ms")
	t.Setenv(envJobLease, "10m")

	got := ResolveJobWorkerSettings()
	if got.Workers != 40 {
		t.Fatalf("workers %d", got.Workers)
	}
	if got.PollInterval != 500*time.Millisecond {
		t.Fatalf("poll %s", got.PollInterval)
	}
	if got.Lease != 10*time.Minute {
		t.Fatalf("lease %s", got.Lease)
	}
}

/*
 * A typo in a deployment must not disable the worker.
 *
 * Zero workers is a queue nothing ever drains; a zero interval is a tight loop;
 * a zero lease lets every replica claim the same job at once. Each of those is
 * worse than ignoring the setting.
 */
func TestNonsenseSizingFallsBackRatherThanBreaking(t *testing.T) {
	for _, value := range []string{"lots", "0", "-4", "5x"} {
		t.Setenv(envJobWorkers, value)
		if got := ResolveJobWorkerSettings(); got.Workers != defaultJobWorkers {
			t.Fatalf("workers=%q gave %d, want the default", value, got.Workers)
		}
	}
	t.Setenv(envJobWorkers, "")

	for _, value := range []string{"soon", "0s", "-1s", "2 seconds"} {
		t.Setenv(envJobPollInterval, value)
		if got := ResolveJobWorkerSettings(); got.PollInterval != defaultJobPollInterval {
			t.Fatalf("poll=%q gave %s, want the default", value, got.PollInterval)
		}
	}
}

/*
 * The guarantee the lease exists to hold.
 *
 * An outbound call may run for 30 seconds. A lease shorter than that lets a
 * second worker claim a job that is still in flight, which is a duplicate
 * service call — the exact thing the idempotency work exists to prevent. The
 * setting is refused rather than the guarantee.
 */
func TestATooShortLeaseIsRefused(t *testing.T) {
	t.Setenv(envJobLease, "10s")
	got := ResolveJobWorkerSettings()
	if got.Lease < 2*time.Minute {
		t.Fatalf("lease %s is short enough to permit a double execution", got.Lease)
	}
}
