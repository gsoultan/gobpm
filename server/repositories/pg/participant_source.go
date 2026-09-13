package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/internal/pkg/configsecret"
	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/repositories/contracts"
	"github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/store/participantsource"
	"github.com/gsoultan/storm"
	"github.com/gsoultan/storm/runtime"
)

type participantSourceRepository struct {
	conn *db.Conn
}

// NewParticipantSourceRepository returns the store of directories a project
// syncs from.
func NewParticipantSourceRepository(conn *db.Conn) contracts.ParticipantSourceRepository {
	return &participantSourceRepository{conn: conn}
}

func (r *participantSourceRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]entities.ParticipantSource, error) {
	ex, err := r.conn.Executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := participantsource.New().
		Where(participantsource.ProjectID.Eq(projectID)).
		Order(participantsource.Name.Asc()).
		All(ctx, ex, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list participant sources: %w", err)
	}
	out := make([]entities.ParticipantSource, 0, len(rows))
	for _, row := range rows {
		out = append(out, sourceFrom(row, true))
	}
	return out, nil
}

// Get returns a source with its credentials masked, for anything that renders
// it.
func (r *participantSourceRepository) Get(ctx context.Context, id uuid.UUID) (entities.ParticipantSource, error) {
	row, err := r.readSource(ctx, id)
	if err != nil {
		return entities.ParticipantSource{}, err
	}
	return sourceFrom(row, true), nil
}

// GetWithSecrets returns a source the syncer can actually use.
//
// A separate method so that handing out a database password is a call somebody
// made on purpose, rather than the default every reader gets.
func (r *participantSourceRepository) GetWithSecrets(ctx context.Context, id uuid.UUID) (entities.ParticipantSource, error) {
	row, err := r.readSource(ctx, id)
	if err != nil {
		return entities.ParticipantSource{}, err
	}
	return sourceFrom(row, false), nil
}

func (r *participantSourceRepository) readSource(ctx context.Context, id uuid.UUID) (participantsource.Row, error) {
	ex, err := r.conn.Executor(ctx)
	if err != nil {
		return participantsource.Row{}, err
	}
	row, found, err := participantsource.New().
		Where(participantsource.ID.Eq(id)).
		One(ctx, ex)
	if err != nil {
		return participantsource.Row{}, fmt.Errorf("could not read participant source: %w", err)
	}
	if !found {
		return participantsource.Row{}, fmt.Errorf("%w: no such participant source", apierr.ErrNotFound)
	}
	return row, nil
}

// Save creates or updates a source, keeping stored credentials the caller did
// not retype.
func (r *participantSourceRepository) Save(ctx context.Context, source entities.ParticipantSource) (uuid.UUID, error) {
	if source.Project == nil || source.Project.ID == uuid.Nil {
		return uuid.Nil, apierr.Invalidf("a source belongs to a project")
	}
	ex, err := r.conn.Executor(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	config := source.Config
	if source.ID != uuid.Nil {
		if stored, err := r.readSource(ctx, source.ID); err == nil {
			// The masking sentinel means "keep what is stored", so editing a
			// query does not require re-typing a database password the browser
			// was never given.
			config = configsecret.Merge(config, decodeConfig(stored.Config))
		}
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not encode the source configuration: %w", err)
	}

	if source.ID == uuid.Nil {
		ins := participantsource.Create()
		ins.SetProjectID(source.Project.ID)
		ins.SetName(source.Name)
		ins.SetKind(source.Kind)
		ins.SetConfig(storm.JSON(encoded))
		ins.SetOnMissing(source.OnMissing)
		ins.SetEnabled(source.Enabled)
		// Set explicitly rather than left to a default: a masked insert omits
		// what nothing assigned, and these columns are NOT NULL with no default
		// because a count is a number rather than an absence. A source that has
		// never run has created nobody, which is zero.
		ins.SetLastRunCreated(0)
		ins.SetLastRunUpdated(0)
		setScheduleOn(ins.SetSchedule, ins.SetScheduleNull, source.Schedule)
		row, err := ins.Insert(ctx, ex)
		if err != nil {
			return uuid.Nil, fmt.Errorf("could not create the source: %w", err)
		}
		return row.ID, nil
	}

	row, err := r.readSource(ctx, source.ID)
	if err != nil {
		return uuid.Nil, err
	}
	mut := participantsource.Mutate(row)
	mut.SetName(source.Name)
	mut.SetKind(source.Kind)
	mut.SetConfig(storm.JSON(encoded))
	mut.SetOnMissing(source.OnMissing)
	mut.SetEnabled(source.Enabled)
	setScheduleOn(mut.SetSchedule, mut.SetScheduleNull, source.Schedule)
	if err := mut.Update(ctx, ex); err != nil {
		return uuid.Nil, fmt.Errorf("could not update the source: %w", err)
	}
	return source.ID, nil
}

// Delete removes a source. The participants it brought in are left alone: they
// are people who exist, not rows belonging to a feed.
func (r *participantSourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ex, err := r.conn.Executor(ctx)
	if err != nil {
		return err
	}
	// The generated Delete marks the row. Stamping deleted_at by hand worked
	// too, and was the thing every table had to remember separately; the model
	// declares it now, so this is the one function that means "remove".
	if err := participantsource.Delete(ctx, ex, id); err != nil {
		if errors.Is(err, runtime.ErrNoRow) {
			return fmt.Errorf("%w: no such participant source", apierr.ErrNotFound)
		}
		return fmt.Errorf("could not delete the source: %w", err)
	}
	return nil
}

// DueForSync returns every enabled, scheduled source across the installation.
//
// Which of them is actually due is decided by the caller from the schedule and
// the last run, because that arithmetic belongs with the schedule vocabulary
// rather than in SQL — and because "due" changes meaning if a run failed.
func (r *participantSourceRepository) DueForSync(ctx context.Context) ([]entities.ParticipantSource, error) {
	ex, err := r.conn.Executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := participantsource.New().
		Where(
			participantsource.Enabled.Eq(true),
			participantsource.Schedule.IsNotNull(),
		).
		All(ctx, ex, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list scheduled sources: %w", err)
	}
	out := make([]entities.ParticipantSource, 0, len(rows))
	for _, row := range rows {
		// With secrets: the worker is going to use them.
		out = append(out, sourceFrom(row, false))
	}
	return out, nil
}

// RecordRun saves how a sync went, so a source that has been failing is visible
// without reading logs.
func (r *participantSourceRepository) RecordRun(ctx context.Context, id uuid.UUID, run entities.SourceRun) error {
	ex, err := r.conn.Executor(ctx)
	if err != nil {
		return err
	}
	row, err := r.readSource(ctx, id)
	if err != nil {
		return err
	}
	mut := participantsource.Mutate(row)
	mut.SetLastRunAt(run.At)
	mut.SetLastRunOk(run.OK)
	mut.SetLastRunCreated(int64(run.Created))
	mut.SetLastRunUpdated(int64(run.Updated))
	if run.Detail == "" {
		mut.SetLastRunDetailNull()
	} else {
		mut.SetLastRunDetail(run.Detail)
	}
	if err := mut.Update(ctx, ex); err != nil {
		return fmt.Errorf("could not record the sync: %w", err)
	}
	return nil
}

func sourceFrom(row participantsource.Row, mask bool) entities.ParticipantSource {
	config := decodeConfig(row.Config)
	if mask {
		config = configsecret.Mask(config)
	}
	source := entities.ParticipantSource{
		ID:        row.ID,
		Project:   &entities.Project{ID: row.ProjectID},
		Name:      row.Name,
		Kind:      row.Kind,
		Config:    config,
		Schedule:  valueOr(row.Schedule),
		OnMissing: row.OnMissing,
		Enabled:   row.Enabled,
	}
	if at, ok := row.LastRunAt.Get(); ok {
		ok2, _ := row.LastRunOk.Get()
		source.LastRun = &entities.SourceRun{
			At:      at,
			OK:      ok2,
			Detail:  valueOr(row.LastRunDetail),
			Created: int(row.LastRunCreated),
			Updated: int(row.LastRunUpdated),
		}
	}
	return source
}

// decodeConfig reads a stored configuration, tolerating one that will not
// parse: a source whose config is corrupt should be visible and fixable rather
// than making the whole list unreadable.
func decodeConfig(raw storm.JSON) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return map[string]any{}
	}
	return config
}

func setScheduleOn(set func(string), setNull func(), schedule string) {
	if schedule == "" {
		setNull()
		return
	}
	set(schedule)
}
