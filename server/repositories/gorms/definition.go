package gorms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/server/repositories/contracts"
	"github.com/gsoultan/metis/server/repositories/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormDefinitionRepository struct {
	db *gorm.DB
}

// NewDefinitionRepository creates a new GORM-based DefinitionRepository.
func NewDefinitionRepository(db *gorm.DB) contracts.DefinitionRepository {
	return &gormDefinitionRepository{db: db}
}

// Get returns a process definition by ID, scoped to the caller's tenant. The
// row carries the deployed BPMN XML, so an unscoped read hands over another
// organization's process model in full.
func (r *gormDefinitionRepository) Get(ctx context.Context, id uuid.UUID) (models.ProcessDefinitionModel, error) {
	var m models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableProcessDefinitions)
	if err := db.First(&m, QualifiedByID(tableProcessDefinitions), id).Error; err != nil {
		return models.ProcessDefinitionModel{}, lookupError(err, "definition")
	}
	return m, nil
}

// GetByKey returns the live version of a definition — the one new instances
// start on — scoped to the caller's tenant. Keys are chosen per project, so two
// organizations can hold the same one; unscoped, this returned whichever
// happened to sort first, which is a leak and a wrong answer at the same time.
//
// The live version is whatever process_definition_releases names, and the
// highest version when it names nothing. That fallback is the upgrade path: an
// installation that has never promoted anything has no release rows and keeps
// the behaviour it had before releases existed. It is also the repair path — a
// release naming a version somebody has since deleted resolves to the highest
// one rather than refusing to start the process at all.
//
// The highest version is read *first*, and its project then decides which
// release applies. Reading the release first would be one query cheaper in the
// staged case and wrong in a way that is hard to see: a key is unique per
// project, not per organization, so two projects in one tenant can both hold
// "expense-approval". Matching a release on the key alone would let one
// project's choice pin the other project's starts. Anchoring to the row this
// lookup would have returned anyway keeps that resolution exactly where it was.
func (r *gormDefinitionRepository) GetByKey(ctx context.Context, key string) (models.ProcessDefinitionModel, error) {
	latest, err := r.GetLatestByKey(ctx, key)
	if err != nil {
		return models.ProcessDefinitionModel{}, err
	}

	release, err := r.GetRelease(ctx, uuid.UUID(latest.ProjectID), key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Nobody has chosen: the highest version is live, as it always was.
			return latest, nil
		}
		return models.ProcessDefinitionModel{}, err
	}
	if release.Version == latest.Version {
		return latest, nil
	}

	live, err := r.getByProjectKeyVersion(ctx, uuid.UUID(latest.ProjectID), key, release.Version)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// The release names a version somebody has deleted. Falling back
			// keeps the process startable; refusing would take a whole process
			// offline over a stale pointer.
			return latest, nil
		}
		return models.ProcessDefinitionModel{}, err
	}
	return live, nil
}

// GetRelease returns which version of key new instances of one project start on
// right now: the newest timeline entry whose activation time has passed.
//
// No such row is a normal answer, not a failure. It means nobody has chosen —
// or that every choice is still in the future — and the reader treats it as "the
// highest version", which is what the engine did before releases existed.
//
// It takes a project rather than resolving by key alone, because a key is unique
// per project and not per organization. A key-only form would answer with some
// other project's choice whenever two projects in one tenant share a key, and
// the answer decides which process model runs.
//
// The clock is read here rather than passed in. A scheduled cutover has to take
// effect for a caller that knows nothing about scheduling — the engine starting
// a process — and threading a "now" through every such path is an invitation to
// pass the wrong one.
func (r *gormDefinitionRepository) GetRelease(ctx context.Context, projectID uuid.UUID, key string) (models.ProcessDefinitionReleaseModel, error) {
	var m models.ProcessDefinitionReleaseModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableDefinitionReleases)
	if err := db.
		Where(QualifiedByProjectID(tableDefinitionReleases), projectID).
		Where(tableDefinitionReleases+"."+ByProcessKey, key).
		Where(tableDefinitionReleases+".activate_at <= ?", time.Now().UTC()).
		Order(tableDefinitionReleases + ".activate_at DESC").
		First(&m).Error; err != nil {
		return models.ProcessDefinitionReleaseModel{}, lookupError(err, "definition release")
	}
	return m, nil
}

// ListReleasesForKey returns one key's whole release timeline, newest first,
// including cutovers that have not happened yet.
//
// The scheduling UI needs the future entries and the effective one in the same
// answer, and the timeline for a single key is a handful of rows.
func (r *gormDefinitionRepository) ListReleasesForKey(ctx context.Context, projectID uuid.UUID, key string) ([]models.ProcessDefinitionReleaseModel, error) {
	var modelsList []models.ProcessDefinitionReleaseModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableDefinitionReleases)
	if err := db.
		Where(QualifiedByProjectID(tableDefinitionReleases), projectID).
		Where(tableDefinitionReleases+"."+ByProcessKey, key).
		Order(tableDefinitionReleases + ".activate_at DESC").
		Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list the release timeline for %s: %w", key, err)
	}
	return modelsList, nil
}

// DeleteScheduledRelease cancels a cutover that has not happened yet.
//
// The activation time is re-read and re-checked inside this call rather than
// trusted from the caller: between rendering a "cancel" button and pressing it,
// the cutover may have taken effect, and deleting it then would silently rewrite
// which version has been in force — and leave running instances on a version the
// timeline no longer admits to having chosen.
//
// Hard delete, because Base's soft delete would leave the row occupying its slot
// in the unique index and refuse a later cutover scheduled for the same instant.
func (r *gormDefinitionRepository) DeleteScheduledRelease(ctx context.Context, projectID uuid.UUID, id uuid.UUID) error {
	db := tenantScopeCondition(ctx, GetTx(ctx, r.db), tableDefinitionReleases)
	result := db.Unscoped().
		Where(QualifiedByID(tableDefinitionReleases), id).
		Where(QualifiedByProjectID(tableDefinitionReleases), projectID).
		Where(tableDefinitionReleases+".activate_at > ?", time.Now().UTC()).
		Delete(&models.ProcessDefinitionReleaseModel{})
	if result.Error != nil {
		return fmt.Errorf("could not cancel the scheduled release: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: no such scheduled release (it may have already taken effect)", apierr.ErrNotFound)
	}
	return nil
}

// getByProjectKeyVersion pins one version within one project, for the same
// reason getReleaseForProject exists: the key alone does not identify a process
// once two projects in a tenant share one.
func (r *gormDefinitionRepository) getByProjectKeyVersion(ctx context.Context, projectID uuid.UUID, key string, version int) (models.ProcessDefinitionModel, error) {
	var m models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableProcessDefinitions)
	if err := db.
		Where(QualifiedByProjectID(tableProcessDefinitions), projectID).
		Where(ByKeyAndVersion(key, version)).
		First(&m).Error; err != nil {
		return models.ProcessDefinitionModel{}, lookupError(err, "definition by key and version")
	}
	return m, nil
}

// GetLatestByKey returns the highest-numbered version of a definition,
// regardless of which one is live.
//
// Deploying needs this and GetByKey cannot serve it: once a version is staged,
// the live one is deliberately not the newest.
func (r *gormDefinitionRepository) GetLatestByKey(ctx context.Context, key string) (models.ProcessDefinitionModel, error) {
	var m models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableProcessDefinitions)
	if err := db.Order(OrderLatestDefinition).Where(ByKey(key)).First(&m).Error; err != nil {
		return models.ProcessDefinitionModel{}, lookupError(err, "definition by key")
	}
	return m, nil
}

// GetByKeyAndVersion pins one version of a definition, scoped to the caller's
// tenant for the same reason GetByKey is.
func (r *gormDefinitionRepository) GetByKeyAndVersion(ctx context.Context, key string, version int) (models.ProcessDefinitionModel, error) {
	var m models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableProcessDefinitions)
	if err := db.Where(ByKeyAndVersion(key, version)).First(&m).Error; err != nil {
		return models.ProcessDefinitionModel{}, lookupError(err, "definition by key and version")
	}
	return m, nil
}

// NextVersion returns the version number a new deployment of key should claim.
//
// See gormDecisionRepository.NextVersion — same contract, same reasoning: the
// count runs over the rows the unique index covers, soft-deleted ones included,
// and the answer is a proposal that the index arbitrates.
func (r *gormDefinitionRepository) NextVersion(ctx context.Context, projectID uuid.UUID, key string) (int, error) {
	db := tenantScopeDB(ctx,
		GetTx(ctx, r.db).Unscoped().Model(&models.ProcessDefinitionModel{}),
		tableProcessDefinitions)

	var highest int
	if err := db.
		Where(QualifiedByProjectID(tableProcessDefinitions), projectID).
		Where(ByKey(key)).
		Select(QueryHighestVersion(tableProcessDefinitions)).
		Scan(&highest).Error; err != nil {
		return 0, fmt.Errorf("could not read the highest definition version: %w", err)
	}
	return highest + 1, nil
}

func (r *gormDefinitionRepository) List(ctx context.Context) ([]models.ProcessDefinitionModel, error) {
	var modelsList []models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), "process_definitions")
	if err := db.Select("process_definitions.id", "process_definitions.project_id", "process_definitions.key", "process_definitions.name", "process_definitions.version", "process_definitions.created_at").Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list definitions: %w", err)
	}
	return modelsList, nil
}

// definitionGraphBatchSize bounds how many definition graphs are held at once
// while scanning. A project keeps every version of every process it has ever
// had, and a graph is the whole BPMN document, so the unbatched form grows
// with installation age — the shape that made BackfillEngineBookkeeping load
// every process instance ever created into memory on every boot.
const definitionGraphBatchSize = 200

// ScanWithGraphs walks the caller's definitions with their node and flow
// graphs hydrated, handing them to visit one batch at a time.
//
// List deliberately projects those columns away, because a list page does not
// need them. A caller that reads what definitions *contain* must ask for them,
// and gets a callback rather than a slice so peak memory is one batch rather
// than the whole installation. Returning an error from visit stops the scan.
func (r *gormDefinitionRepository) ScanWithGraphs(ctx context.Context, visit func([]models.ProcessDefinitionModel) error) error {
	var batch []models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), "process_definitions")

	// FindInBatches orders by primary key and pages internally, so the scan is
	// stable and never materializes the full set.
	result := db.FindInBatches(&batch, definitionGraphBatchSize, func(*gorm.DB, int) error {
		return visit(batch)
	})
	if result.Error != nil {
		return fmt.Errorf("could not scan definitions with graphs: %w", result.Error)
	}
	return nil
}

// ListByProjectPaged returns one page of a project's definitions, newest first.
//
// The order is table-qualified: tenant scoping joins the projects table, which
// carries a created_at of its own, and a bare column name is ambiguous the
// moment that join is present — which is every request-driven call.
func (r *gormDefinitionRepository) ListByProjectPaged(ctx context.Context, projectID uuid.UUID, p contracts.Pagination) (contracts.Page[models.ProcessDefinitionModel], error) {
	base := tenantScopeDB(ctx, GetTx(ctx, r.db), "process_definitions").
		Model(&models.ProcessDefinitionModel{}).
		Where("process_definitions.project_id = ?", projectID)
	return countAndPage[models.ProcessDefinitionModel](base, p, "process_definitions.created_at DESC")
}

func (r *gormDefinitionRepository) Create(ctx context.Context, m models.ProcessDefinitionModel) error {
	// Refuse a process definition planted in another organization's project.
	if err := requireProjectInTenant(ctx, GetTx(ctx, r.db), uuid.UUID(m.ProjectID)); err != nil {
		return err
	}
	if err := GetTx(ctx, r.db).Create(&m).Error; err != nil {
		return fmt.Errorf("could not create definition: %w", err)
	}
	return nil
}

// tableProcessDefinitions is the SQL table behind ProcessDefinitionModel,
// needed by name so the tenant scope can build its clauses.
const tableProcessDefinitions = "process_definitions"

// Delete removes a process definition, refusing an ID outside the caller's
// tenant.
func (r *gormDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := GetTx(ctx, r.db)
	if err := requireVisibleToTenant(ctx, db, tableProcessDefinitions, &models.ProcessDefinitionModel{}, id); err != nil {
		return err
	}
	if err := db.Delete(&models.ProcessDefinitionModel{}, QualifiedByID(tableProcessDefinitions), id).Error; err != nil {
		return fmt.Errorf("could not delete definition: %w", err)
	}
	return nil
}

func (r *gormDefinitionRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.ProcessDefinitionModel, error) {
	var modelsList []models.ProcessDefinitionModel
	if err := GetTx(ctx, r.db).Select("id", "project_id", "key", "name", "version", "created_at").Where(QueryByProjectID, projectID).Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list definitions by project: %w", err)
	}
	return modelsList, nil
}

// tableDefinitionReleases is the SQL table behind ProcessDefinitionReleaseModel,
// needed by name so the tenant scope can build its clauses.
const tableDefinitionReleases = "process_definition_releases"

// ByProcessKey builds a condition on the release table's key column.
//
// Spelled process_key rather than key so it can be a plain string condition:
// `key` is reserved on MySQL and would force the map form every other query on
// that column has to use. See models.ProcessDefinitionReleaseModel.
const ByProcessKey = "process_key = ?"

// ScheduleRelease adds a timeline entry: from activateAt, key runs version.
//
// A time in the past (or now) is a promotion that takes effect immediately; a
// time in the future is an arranged cutover. There is no separate call for the
// two, because they differ only in the number.
//
// An upsert on the unique triple rather than a read-then-write: two
// administrators arranging the same instant would both read "no row" and both
// insert, and the loser would be told the database refused a duplicate key
// rather than being given the answer they asked for.
func (r *gormDefinitionRepository) ScheduleRelease(ctx context.Context, projectID uuid.UUID, key string, version int, activateAt time.Time) error {
	// Refuse a release written into another organization's project, for the same
	// reason Create does: project_id is caller-supplied.
	if err := requireProjectInTenant(ctx, GetTx(ctx, r.db), projectID); err != nil {
		return err
	}
	m := models.ProcessDefinitionReleaseModel{
		ProjectID:  models.UUID(projectID),
		ProcessKey: key,
		Version:    version,
		// Stored in UTC so that comparing it against the clock does not depend on
		// where the row was written or which replica reads it back.
		ActivateAt: activateAt.UTC(),
	}
	if err := GetTx(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "process_key"}, {Name: "project_id"}, {Name: "activate_at"}},
		DoUpdates: clause.AssignmentColumns([]string{"version", "updated_at"}),
	}).Create(&m).Error; err != nil {
		return fmt.Errorf("could not schedule the live version of %s: %w", key, err)
	}
	return nil
}

// ListReleases returns every key's live version for one project, so a caller
// rendering a list of processes can mark them in one query rather than one per
// row.
func (r *gormDefinitionRepository) ListReleases(ctx context.Context, projectID uuid.UUID) ([]models.ProcessDefinitionReleaseModel, error) {
	var modelsList []models.ProcessDefinitionReleaseModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableDefinitionReleases)
	if err := db.
		Where(QualifiedByProjectID(tableDefinitionReleases), projectID).
		Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list definition releases: %w", err)
	}
	return modelsList, nil
}

// ListVersionsByKey returns every version deployed under one key, newest first.
//
// ListByProject would serve this with a filter, but it returns every version of
// every process the project has ever had — a list that only grows — and the
// caller here wants one process's history.
func (r *gormDefinitionRepository) ListVersionsByKey(ctx context.Context, projectID uuid.UUID, key string) ([]models.ProcessDefinitionModel, error) {
	var modelsList []models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableProcessDefinitions)
	if err := db.
		Select("process_definitions.id", "process_definitions.project_id", "process_definitions.key",
			"process_definitions.name", "process_definitions.version", "process_definitions.created_at").
		Where(QualifiedByProjectID(tableProcessDefinitions), projectID).
		Where(ByKey(key)).
		Order(OrderLatestDefinition).
		Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("could not list versions of %s: %w", key, err)
	}
	return modelsList, nil
}

// GetLiveByProjectKey returns the version of key that new instances of one
// project start on.
//
// This is the form the engine uses, and it exists because a process key is
// unique per *project*, not per organization. GetByKey resolves by key alone —
// all a caller holding only a key can do — but that makes two projects in one
// tenant sharing "expense-approval" ambiguous, and the ambiguity decides which
// process model actually runs. The engine always knows the project, so it never
// has to guess.
func (r *gormDefinitionRepository) GetLiveByProjectKey(ctx context.Context, projectID uuid.UUID, key string) (models.ProcessDefinitionModel, error) {
	if projectID == uuid.Nil {
		return r.GetByKey(ctx, key)
	}

	release, err := r.GetRelease(ctx, projectID, key)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ProcessDefinitionModel{}, err
	}
	if err == nil {
		live, err := r.getByProjectKeyVersion(ctx, projectID, key, release.Version)
		if err == nil {
			return live, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ProcessDefinitionModel{}, err
		}
		// The release names a version somebody deleted; fall through to the
		// highest rather than take the process offline over a stale pointer.
	}

	var m models.ProcessDefinitionModel
	db := tenantScopeDB(ctx, GetTx(ctx, r.db), tableProcessDefinitions)
	if err := db.
		Where(QualifiedByProjectID(tableProcessDefinitions), projectID).
		Where(ByKey(key)).
		Order(OrderLatestDefinition).
		First(&m).Error; err != nil {
		return models.ProcessDefinitionModel{}, lookupError(err, "definition by key")
	}
	return m, nil
}

// GetByProjectKeyAndVersion pins one version within one project, for the same
// reason GetLiveByProjectKey exists.
func (r *gormDefinitionRepository) GetByProjectKeyAndVersion(ctx context.Context, projectID uuid.UUID, key string, version int) (models.ProcessDefinitionModel, error) {
	if projectID == uuid.Nil {
		return r.GetByKeyAndVersion(ctx, key, version)
	}
	return r.getByProjectKeyVersion(ctx, projectID, key, version)
}
