package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/server/repositories/contracts"
	"github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/models"
	"github.com/gsoultan/metis/server/repositories/store/form"
	"github.com/gsoultan/storm/runtime"
)

type formRepository struct{ conn }

// NewFormRepository returns the store of task forms.
func NewFormRepository(c *db.Conn) contracts.FormRepository {
	return &formRepository{conn{conn: c}}
}

func (r *formRepository) Create(ctx context.Context, f models.FormModel) error {
	projectID := uuid.UUID(f.ProjectID)
	if err := r.requireProjectInTenant(ctx, projectID); err != nil {
		return err
	}
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return err
	}
	fields, err := jsonOf(f.Schema)
	if err != nil {
		return fmt.Errorf("could not encode the form: %w", err)
	}

	ins := form.Create()
	if id := uuid.UUID(f.ID); id != uuid.Nil {
		ins.SetID(id)
	}
	ins.SetProjectID(projectID)
	ins.SetKey(f.Key)
	ins.SetName(f.Name)
	ins.SetFields(fields)
	if _, err := ins.Insert(ctx, ex); err != nil {
		return fmt.Errorf("could not create the form: %w", err)
	}
	return nil
}

func (r *formRepository) Get(ctx context.Context, id uuid.UUID) (models.FormModel, error) {
	row, err := r.one(ctx, form.ID.Eq(id))
	if err != nil {
		return models.FormModel{}, err
	}
	return formFrom(row)
}

func (r *formRepository) GetByKey(ctx context.Context, projectID uuid.UUID, key string) (models.FormModel, error) {
	row, err := r.one(ctx, form.ProjectID.Eq(projectID), form.Key.Eq(key))
	if err != nil {
		return models.FormModel{}, err
	}
	return formFrom(row)
}

func (r *formRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.FormModel, error) {
	scoped, visible, err := r.scopedProjects(ctx, projectID)
	if err != nil || !visible {
		return nil, err
	}
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return nil, err
	}
	q := form.New().Order(form.Name.Asc())
	if scoped != nil {
		q = q.Where(form.ProjectID.In(uuidsToRaw(scoped)...))
	}
	rows, err := q.All(ctx, ex, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list forms: %w", err)
	}
	out := make([]models.FormModel, 0, len(rows))
	for _, row := range rows {
		f, err := formFrom(row)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

func (r *formRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Read first, so the scope is checked before the row is touched: a delete
	// that refused only after removing it would be no refusal at all.
	if _, err := r.one(ctx, form.ID.Eq(id)); err != nil {
		return err
	}
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return err
	}
	if err := form.Delete(ctx, ex, id); err != nil {
		if errors.Is(err, runtime.ErrNoRow) {
			return fmt.Errorf("%w: no such form", apierr.ErrNotFound)
		}
		return fmt.Errorf("could not delete the form: %w", err)
	}
	return nil
}

// one reads a single form within the caller's scope.
func (r *formRepository) one(ctx context.Context, preds ...form.Pred) (form.Row, error) {
	scope, err := r.scopeOf(ctx)
	if err != nil {
		return form.Row{}, err
	}
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return form.Row{}, err
	}
	q := form.New().Where(preds...)
	if !scope.unrestricted() {
		if len(scope.projects) == 0 {
			return form.Row{}, fmt.Errorf("%w: no such form", apierr.ErrNotFound)
		}
		q = q.Where(form.ProjectID.In(uuidsToRaw(scope.projects)...))
	}
	row, found, err := q.One(ctx, ex)
	if err != nil {
		return form.Row{}, fmt.Errorf("could not read the form: %w", err)
	}
	if !found {
		return form.Row{}, fmt.Errorf("%w: no such form", apierr.ErrNotFound)
	}
	return row, nil
}

func formFrom(row form.Row) (models.FormModel, error) {
	schema, err := mapOf(row.Fields)
	if err != nil {
		return models.FormModel{}, fmt.Errorf("could not decode a form: %w", err)
	}
	return models.FormModel{
		Base: models.Base{
			ID:        models.UUID(row.ID),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		ProjectID: models.UUID(row.ProjectID),
		Key:       row.Key,
		Name:      row.Name,
		Schema:    schema,
	}, nil
}
