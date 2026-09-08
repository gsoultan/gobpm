package impl

import (
	"context"
	"sync"
	"time"

	"github.com/gsoultan/metis/internal/pkg/envvar"
	"github.com/rs/zerolog/log"
)

// defaultShutdownDrain is how long a stopping process waits for the jobs it has
// already claimed to finish.
//
// Long enough for an ordinary service task — the outbound HTTP client's own
// budget is 30s, so this covers the common case — and short enough to sit
// inside a container runtime's termination grace period, which is 30s in the
// shipped Kubernetes manifest.
const defaultShutdownDrain = 20 * time.Second

const envShutdownDrain = "METIS_SHUTDOWN_DRAIN"

// workerLifecycle tracks the goroutines the job worker owns so that a stopping
// process can wait for them.
//
// Without it, shutdown cancelled the context and returned immediately. Anything
// already running was abandoned mid-flight: its final status write used the
// cancelled context and failed, so the row stayed marked running with this
// worker's lock on it until the five-minute lease expired. On the shipped
// manifest — Recreate strategy, so every rollout is a full stop — that meant
// every deploy left claimed work frozen for five minutes, and an operator
// watching the incident inbox saw jobs that were not moving and no reason why.
type workerLifecycle struct {
	mu      sync.Mutex
	running sync.WaitGroup
	stopped bool
	// inFlight answers "is anything running" without waiting. The WaitGroup
	// alone cannot: reading it means starting a goroutine to Wait on, which may
	// not have been scheduled yet, so a zero budget reported work in flight on
	// an idle worker.
	inFlight int
}

// ShutdownDrain is how long to wait for in-flight jobs.
func ShutdownDrain() time.Duration {
	if raw := envvar.Get(envShutdownDrain); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d >= 0 {
			return d
		}
	}
	return defaultShutdownDrain
}

// begin registers a job about to run. It reports false once stopping has begun,
// so a tick that fires during shutdown does not claim more work.
func (w *workerLifecycle) begin() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return false
	}
	w.running.Add(1)
	w.inFlight++
	return true
}

func (w *workerLifecycle) done() {
	w.mu.Lock()
	w.inFlight--
	w.mu.Unlock()
	w.running.Done()
}

// idle reports whether nothing is running right now.
func (w *workerLifecycle) idle() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.inFlight == 0
}

// stopping marks the worker as no longer accepting work.
func (w *workerLifecycle) stopping() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stopped = true
}

// wait blocks until every in-flight job has finished or the budget runs out.
// It reports whether the drain completed.
func (w *workerLifecycle) wait(budget time.Duration) bool {
	// Answered without waiting, so a zero budget is not a race against the
	// scheduler starting the goroutine below.
	if w.idle() {
		return true
	}
	if budget <= 0 {
		return false
	}

	finished := make(chan struct{})
	go func() {
		w.running.Wait()
		close(finished)
	}()

	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case <-finished:
		return true
	case <-timer.C:
		return false
	}
}

// StopWorkers stops claiming new jobs and waits for the claimed ones.
//
// Called during shutdown, before the database handle is closed. A job that does
// not finish inside the budget is left to its lease, which is the pre-existing
// behaviour and the reason the budget exists at all: waiting forever would turn
// one stuck partner call into a pod that will not terminate.
func (s *jobService) StopWorkers(ctx context.Context) error {
	s.lifecycle.stopping()

	budget := ShutdownDrain()
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < budget {
			budget = remaining
		}
	}

	start := time.Now()
	if s.lifecycle.wait(budget) {
		log.Info().Dur("took", time.Since(start)).Msg("In-flight jobs finished; worker stopped")
		return nil
	}

	// Said out loud rather than swallowed: this is the case where a deploy
	// leaves work frozen until its lease expires, and somebody watching a stuck
	// instance afterwards deserves to find the reason in the logs.
	log.Warn().
		Dur("waited", time.Since(start)).
		Dur("budget", budget).
		Msg("Some jobs were still running at shutdown; they will be retried after their lock expires. " +
			"Raise METIS_SHUTDOWN_DRAIN if this happens on every deploy.")
	return nil
}

// detach returns a context that carries the parent's values but not its
// cancellation, with a short budget of its own.
//
// A job's final status write must not be made on the context that was just
// cancelled: it fails, the row keeps this worker's lock, and the work is frozen
// until the lease expires. The values are kept because the write still needs
// the system marker the worker put there — without it, the strict tenant scope
// would refuse the update.
func detach(ctx context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), budget)
}
