package pg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/server/repositories/contracts"
	"github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/models"
	"github.com/gsoultan/metis/server/repositories/store/deployment"
	"github.com/gsoultan/metis/server/repositories/store/resource"
)

type deploymentRepository struct{ conn }

// NewDeploymentRepository returns the record of what was deployed together.
//
// A deployment is the unit somebody uploaded: several BPMN and DMN files that
// were meant to go live at once. The definitions parsed out of it are separate
// rows with their own versions; this is what says they arrived together.
func NewDeploymentRepository(c *db.Conn) contracts.DeploymentRepository {
	return &deploymentRepository{conn{conn: c}}
}

// Create records a deployment and the files it carried.
//
// One transaction, because a deployment with none of its resources is a record
// of an upload nobody can inspect — and the reason to keep the files at all is
// so somebody can see what was actually deployed.
func (r *deploymentRepository) Create(ctx context.Context, d models.DeploymentModel) error {
	projectID := uuid.UUID(d.ProjectID)
	if err := r.requireProjectInTenant(ctx, projectID); err != nil {
		return err
	}
	return r.conn.conn.Transact(ctx, func(txCtx context.Context) error {
		ex, err := r.conn.conn.Executor(txCtx)
		if err != nil {
			return err
		}
		ins := deployment.Create()
		if id := uuid.UUID(d.ID); id != uuid.Nil {
			ins.SetID(id)
		}
		ins.SetProjectID(projectID)
		ins.SetName(d.Name)
		row, err := ins.Insert(txCtx, ex)
		if err != nil {
			return fmt.Errorf("could not create the deployment: %w", err)
		}

		for _, file := range d.Resources {
			res := resource.Create()
			if id := uuid.UUID(file.ID); id != uuid.Nil {
				res.SetID(id)
			}
			res.SetDeploymentID(row.ID)
			res.SetName(file.Name)
			res.SetContent(file.Content)
			res.SetType(file.Type)
			if _, err := res.Insert(txCtx, ex); err != nil {
				return fmt.Errorf("could not record the deployed file %q: %w", file.Name, err)
			}
		}
		return nil
	})
}

func (r *deploymentRepository) Get(ctx context.Context, id uuid.UUID) (models.DeploymentModel, error) {
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return models.DeploymentModel{}, err
	}
	row, found, err := deployment.New().Where(deployment.ID.Eq(id)).One(ctx, ex)
	if err != nil {
		return models.DeploymentModel{}, fmt.Errorf("could not read the deployment: %w", err)
	}
	if !found {
		return models.DeploymentModel{}, fmt.Errorf("%w: no such deployment", apierr.ErrNotFound)
	}
	if err := r.requireProjectInTenant(ctx, uuid.UUID(row.ProjectID)); err != nil {
		return models.DeploymentModel{}, fmt.Errorf("%w: no such deployment", apierr.ErrNotFound)
	}
	d := deploymentFrom(row)
	files, err := r.ListResources(ctx, id)
	if err != nil {
		return models.DeploymentModel{}, err
	}
	d.Resources = files
	return d, nil
}

func (r *deploymentRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.DeploymentModel, error) {
	scoped, visible, err := r.scopedProjects(ctx, projectID)
	if err != nil || !visible {
		return nil, err
	}
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return nil, err
	}
	q := deployment.New().Order(deployment.CreatedAt.Desc())
	if scoped != nil {
		q = q.Where(deployment.ProjectID.In(uuidsToRaw(scoped)...))
	}
	rows, err := q.All(ctx, ex, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list deployments: %w", err)
	}
	// Without their files: a project's deployment list is a list, and loading
	// every uploaded file to render it would read megabytes to show names.
	out := make([]models.DeploymentModel, 0, len(rows))
	for _, row := range rows {
		out = append(out, deploymentFrom(row))
	}
	return out, nil
}

func (r *deploymentRepository) GetResource(ctx context.Context, id uuid.UUID) (models.ResourceModel, error) {
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return models.ResourceModel{}, err
	}
	row, found, err := resource.New().Where(resource.ID.Eq(id)).One(ctx, ex)
	if err != nil {
		return models.ResourceModel{}, fmt.Errorf("could not read the deployed file: %w", err)
	}
	if !found {
		return models.ResourceModel{}, fmt.Errorf("%w: no such deployed file", apierr.ErrNotFound)
	}
	// Scoped through its deployment: a file carries no project of its own.
	if err := r.deploymentVisible(ctx, uuid.UUID(row.DeploymentID)); err != nil {
		return models.ResourceModel{}, fmt.Errorf("%w: no such deployed file", apierr.ErrNotFound)
	}
	return resourceFrom(row), nil
}

// ListResources returns the files one deployment carried.
//
// Scoped through the deployment, and answered with nothing rather than a
// refusal: a list's contract is rows, and every caller already handles finding
// none. Get is where naming somebody else's deployment is an error.
func (r *deploymentRepository) ListResources(ctx context.Context, deploymentID uuid.UUID) ([]models.ResourceModel, error) {
	if !r.canSeeDeployment(ctx, deploymentID) {
		return nil, nil
	}
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := resource.New().
		Where(resource.DeploymentID.Eq(deploymentID)).
		Order(resource.Name.Asc()).
		All(ctx, ex, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list the deployed files: %w", err)
	}
	out := make([]models.ResourceModel, 0, len(rows))
	for _, row := range rows {
		out = append(out, resourceFrom(row))
	}
	return out, nil
}

func deploymentFrom(row deployment.Row) models.DeploymentModel {
	return models.DeploymentModel{
		Base: models.Base{
			ID:        models.UUID(row.ID),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		ProjectID: models.UUID(row.ProjectID),
		Name:      row.Name,
	}
}

func resourceFrom(row resource.Row) models.ResourceModel {
	return models.ResourceModel{
		Base: models.Base{
			ID:        models.UUID(row.ID),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		DeploymentID: models.UUID(row.DeploymentID),
		Name:         row.Name,
		Content:      row.Content,
		Type:         row.Type,
	}
}

// canSeeDeployment answers the visibility question as a boolean.
//
// A bool rather than an error because the caller turns "no" into an empty list,
// and a function that returns an error the caller must discard reads like a
// mistake — it is one, often enough that the linter refuses it.
func (r *deploymentRepository) canSeeDeployment(ctx context.Context, deploymentID uuid.UUID) bool {
	return r.deploymentVisible(ctx, deploymentID) == nil
}

// deploymentVisible reports whether a deployment is one the caller may see.
//
// Separate from Get because Get loads the files, and checking visibility is a
// question about one row rather than a reason to read a megabyte of BPMN.
func (r *deploymentRepository) deploymentVisible(ctx context.Context, deploymentID uuid.UUID) error {
	ex, err := r.conn.conn.Executor(ctx)
	if err != nil {
		return err
	}
	row, found, err := deployment.New().Where(deployment.ID.Eq(deploymentID)).One(ctx, ex)
	if err != nil {
		return fmt.Errorf("could not read the deployment: %w", err)
	}
	if !found {
		return fmt.Errorf("%w: no such deployment", apierr.ErrNotFound)
	}
	if err := r.requireProjectInTenant(ctx, uuid.UUID(row.ProjectID)); err != nil {
		return fmt.Errorf("%w: no such deployment", apierr.ErrNotFound)
	}
	return nil
}
