package impl

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/repositories/models"

	"github.com/google/uuid"
	servicecontracts "github.com/gsoultan/metis/server/domains/services/contracts"
	"github.com/gsoultan/metis/server/repositories"
)

type migrationService struct {
	repo repositories.Repository
}

// NewMigrationService creates a new MigrationService implementation.
func NewMigrationService(
	repo repositories.Repository,
) servicecontracts.MigrationService {
	return &migrationService{
		repo: repo,
	}
}

// MigrateInstances moves running instances from one version of a process onto
// another, rewriting the nodes they are parked on through nodeMapping.
//
// Nothing reaches this over HTTP or gRPC — the supported way to change version
// is to promote a new one and let the old one drain, because an instance that
// finishes on the graph it started with cannot be broken by an edit. This exists
// for the case drain cannot serve: work already in flight on a version that must
// not continue.
//
// It refuses far more than it used to. The first version validated nothing: a
// mapping to a node the target did not have was accepted and applied, migrating
// across two unrelated process keys was accepted, and a token left on a node the
// target lacked hung the request forever the next time anybody advanced it.
// Everything below is checked *before* a single row is written, because a
// half-migrated instance is worse than a refused migration — the caller can retry
// a refusal, and cannot un-strand a token.
func (s *migrationService) MigrateInstances(ctx context.Context, sourceDefID uuid.UUID, targetDefID uuid.UUID, nodeMapping map[string]string) error {
	// Every refusal is worked out by the planner, so what a dry run showed and
	// what an apply does cannot disagree. Two copies of these checks is how a
	// preview comes to say "this is fine" about something the apply rejects.
	plan, err := s.PlanInstanceMigration(ctx, sourceDefID, targetDefID, nodeMapping)
	if err != nil {
		return err
	}
	if !plan.Applicable() {
		return apierr.Invalidf("%s", strings.Join(plan.Refusals, "; "))
	}
	if plan.Instances == 0 {
		return nil
	}
	return s.apply(ctx, sourceDefID, targetDefID, nodeMapping)
}

// PlanInstanceMigration works out what MigrateInstances would do, and writes
// nothing.
//
// Refusals are collected rather than returned one at a time: somebody fixing a
// node mapping wants the whole list, not to rediscover the next problem after
// each correction.
func (s *migrationService) PlanInstanceMigration(ctx context.Context, sourceDefID uuid.UUID, targetDefID uuid.UUID, nodeMapping map[string]string) (entities.MigrationPlan, error) {
	var plan entities.MigrationPlan
	if sourceDefID == targetDefID {
		return plan, apierr.Invalidf("the source and target versions are the same definition")
	}

	source, err := s.repo.Definition().Get(ctx, sourceDefID)
	if err != nil {
		return plan, fmt.Errorf("source definition: %w", err)
	}
	target, err := s.repo.Definition().Get(ctx, targetDefID)
	if err != nil {
		return plan, fmt.Errorf("target definition: %w", err)
	}
	plan.SourceKey = source.Key
	plan.SourceVersion = source.Version
	plan.TargetVersion = target.Version
	plan.TargetID = targetDefID

	// Same process, or it is not a version change. Migrating "expense-approval"
	// onto "supplier-onboarding" was previously accepted and left every instance
	// claiming to be running a process it had never started.
	if source.Key != target.Key {
		return plan, apierr.Invalidf("cannot migrate %q onto %q: they are different processes, not two versions of one",
			source.Key, target.Key)
	}
	if source.ProjectID != target.ProjectID {
		return plan, apierr.Invalidf("cannot migrate between projects")
	}

	targetNodes := make(map[string]struct{}, len(target.Nodes))
	for _, node := range target.Nodes {
		targetNodes[node.ID] = struct{}{}
	}

	// A mapping that names a node the target does not have is a typo that would
	// park a token somewhere unreachable. Checked first because it is wrong
	// regardless of which instances happen to be running.
	var badTargets []string
	for from, to := range nodeMapping {
		if _, ok := targetNodes[to]; !ok {
			badTargets = append(badTargets, fmt.Sprintf("%s→%s", from, to))
		}
	}
	if len(badTargets) > 0 {
		slices.Sort(badTargets)
		// Raised rather than collected: a mapping that names a node the target
		// does not have is wrong on its face, independently of what is running,
		// so there is nothing further to plan.
		return plan, apierr.Invalidf("version %d of %q has no node for %s",
			target.Version, target.Key, strings.Join(badTargets, ", "))
	}

	instances, err := s.repo.Process().ListByDefinition(ctx, sourceDefID)
	if err != nil {
		return plan, fmt.Errorf("failed to list instances for migration: %w", err)
	}
	plan.Instances = len(instances)
	if len(instances) == 0 {
		return plan, nil
	}

	// Every place the instances currently sit has to land on a node the target
	// actually has — mapped there explicitly, or carried over because the target
	// still has a node of that name. An unlisted one is exactly the token the
	// old code stranded.
	moves, unlandable, err := s.survey(ctx, instances, targetNodes, nodeMapping)
	if err != nil {
		return plan, err
	}
	plan.Moves = moves
	if len(unlandable) > 0 {
		plan.Refusals = append(plan.Refusals, fmt.Sprintf(
			"version %d of %q has nowhere to put the work parked on %s; map each one to a node it does have",
			target.Version, target.Key, strings.Join(unlandable, ", ")))
	}
	return plan, nil
}

// apply performs the migration. Every refusal has already been made by the
// planner, so this only writes.
func (s *migrationService) apply(ctx context.Context, sourceDefID, targetDefID uuid.UUID, nodeMapping map[string]string) error {
	instances, err := s.repo.Process().ListByDefinition(ctx, sourceDefID)
	if err != nil {
		return fmt.Errorf("failed to list instances for migration: %w", err)
	}

	for _, instance := range instances {
		err := s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
			instance.DefinitionID = models.UUID(targetDefID)
			for i := range instance.Tokens {
				instance.Tokens[i].NodeID = mapNode(nodeMapping, instance.Tokens[i].NodeID)
			}
			if err := s.repo.Process().Update(txCtx, instance); err != nil {
				return err
			}

			tasks, err := s.repo.Task().ListByInstance(txCtx, uuid.UUID(instance.ID))
			if err != nil {
				return err
			}
			for _, task := range tasks {
				mapped := mapNode(nodeMapping, task.NodeID)
				if mapped == task.NodeID {
					continue
				}
				task.NodeID = mapped
				if err := s.repo.Task().Update(txCtx, task); err != nil {
					return err
				}
			}

			jobs, err := s.repo.Job().ListByInstance(txCtx, uuid.UUID(instance.ID))
			if err != nil {
				return err
			}
			for _, job := range jobs {
				job.DefinitionID = models.UUID(targetDefID)
				job.NodeID = mapNode(nodeMapping, job.NodeID)
				if err := s.repo.Job().Update(txCtx, job); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to migrate instance %s: %w", instance.ID, err)
		}
	}

	return nil
}

// survey works out where the running work sits and where it would land.
//
// One pass answers both questions the caller has: the plan to show, and the
// nodes that have nowhere to go. Doing them separately meant walking every
// instance's tasks and jobs twice — and, worse, meant a preview and an apply
// could disagree about what "landable" meant.
//
// Moves are aggregated by node rather than listed per instance: a cutover of a
// thousand purchase orders parked on three nodes is three lines, not a thousand.
func (s *migrationService) survey(
	ctx context.Context,
	instances []models.ProcessInstanceModel,
	targetNodes map[string]struct{},
	nodeMapping map[string]string,
) (moves []entities.NodeMove, unlandable []string, err error) {
	type counts struct{ tokens, tasks, jobs int }
	byNode := map[string]*counts{}
	stranded := map[string]struct{}{}

	count := func(nodeID string, add func(*counts)) {
		if nodeID == "" {
			return
		}
		if _, ok := targetNodes[mapNode(nodeMapping, nodeID)]; !ok {
			stranded[nodeID] = struct{}{}
		}
		c, ok := byNode[nodeID]
		if !ok {
			c = &counts{}
			byNode[nodeID] = c
		}
		add(c)
	}

	for _, instance := range instances {
		for _, token := range instance.Tokens {
			count(token.NodeID, func(c *counts) { c.tokens++ })
		}
		// Tasks and jobs are separate rows pointing at the same graph. A task
		// left on a node the target lacks is work nobody can complete, and a job
		// is a timer or a service call that fires into nothing.
		tasks, listErr := s.repo.Task().ListByInstance(ctx, uuid.UUID(instance.ID))
		if listErr != nil {
			return nil, nil, listErr
		}
		for _, task := range tasks {
			if task.Status == models.TaskUnclaimed || task.Status == models.TaskClaimed || task.Status == models.TaskDelegated {
				count(task.NodeID, func(c *counts) { c.tasks++ })
			}
		}
		jobs, listErr := s.repo.Job().ListByInstance(ctx, uuid.UUID(instance.ID))
		if listErr != nil {
			return nil, nil, listErr
		}
		for _, job := range jobs {
			count(job.NodeID, func(c *counts) { c.jobs++ })
		}
	}

	for from, c := range byNode {
		to, mapped := nodeMapping[from]
		if !mapped {
			to = from
		}
		moves = append(moves, entities.NodeMove{
			From:   from,
			To:     to,
			Tokens: c.tokens,
			Tasks:  c.tasks,
			Jobs:   c.jobs,
			Mapped: mapped,
		})
	}
	// Sorted so the same plan reads the same way twice; a map's order would make
	// a preview look different every time it was refreshed.
	slices.SortFunc(moves, func(a, b entities.NodeMove) int { return strings.Compare(a.From, b.From) })

	unlandable = make([]string, 0, len(stranded))
	for nodeID := range stranded {
		unlandable = append(unlandable, nodeID)
	}
	slices.Sort(unlandable)
	return moves, unlandable, nil
}

func mapNode(nodeMapping map[string]string, nodeID string) string {
	if mapped, ok := nodeMapping[nodeID]; ok {
		return mapped
	}
	return nodeID
}
