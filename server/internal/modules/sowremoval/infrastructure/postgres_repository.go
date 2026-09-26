package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	abortiondomain "server/internal/modules/abortion/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
	sowremovaldomain "server/internal/modules/sowremoval/domain"
	"server/internal/modules/sowremoval/ports"
)

var ErrSowRemovalNotFound = ports.ErrSowRemovalNotFound

type PostgresSowRemovalRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSowRemovalRepository creates a repository backed by a pgx pool.
// It stores the pool used for all removal, sow, service and abortion queries.
func NewPostgresSowRemovalRepository(pool *pgxpool.Pool) *PostgresSowRemovalRepository {
	return &PostgresSowRemovalRepository{pool: pool}
}

// GetSow fetches the sow referenced by a removal by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresSowRemovalRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
	result := &sowdomain.Sow{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, code, location, active, entry_date, birth_date, note, state,
		       origin, parity, breed_id, created_at, updated_at, created_by, updated_by
		FROM sows
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&result.Code,
		&result.Location,
		&result.Active,
		&result.EntryDate,
		&result.BirthDate,
		&result.Note,
		&result.State,
		&result.Origin,
		&result.Parity,
		&result.BreedID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrSowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar la cerda: %w", err)
	}
	return result, nil
}

// GetLastService fetches the most recent service of a sow with its mounts.
// It returns ports.ErrServiceNotFound when the sow has no service.
func (r *PostgresSowRemovalRepository) GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error) {
	result := &servicedomain.Service{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE sow_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, sowID).Scan(
		&result.ID,
		&result.SowID,
		&result.ExpectedFarrowingDate,
		&result.Note,
		&result.State,
		&result.Location,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrServiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el servicio: %w", err)
	}

	result.Mounts, err = r.listMounts(ctx, result.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetLastAbortion fetches the most recent abortion of a sow.
// It returns ports.ErrAbortionNotFound when the sow has no abortion.
func (r *PostgresSowRemovalRepository) GetLastAbortion(ctx context.Context, sowID uuid.UUID) (*abortiondomain.Abortion, error) {
	result := &abortiondomain.Abortion{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, sow_id, service_id, abortion_date, cause, note,
		       created_at, updated_at, created_by, updated_by
		FROM abortions
		WHERE sow_id = $1
		ORDER BY abortion_date DESC, created_at DESC
		LIMIT 1
	`, sowID).Scan(
		&result.ID,
		&result.SowID,
		&result.ServiceID,
		&result.AbortionDate,
		&result.Cause,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrAbortionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el aborto: %w", err)
	}
	return result, nil
}

// listMounts fetches every mount of a service ordered by mount number.
// It returns an empty slice when the service has no mounts.
func (r *PostgresSowRemovalRepository) listMounts(ctx context.Context, serviceID uuid.UUID) ([]*servicedomain.Mount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, service_id, boar_id, operator_id, mount_number, mount_date,
		       type, note, created_at, updated_at, created_by, updated_by
		FROM mounts
		WHERE service_id = $1
		ORDER BY mount_number ASC
	`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}
	defer rows.Close()

	mounts := make([]*servicedomain.Mount, 0)
	for rows.Next() {
		result := &servicedomain.Mount{}
		if err := rows.Scan(
			&result.ID,
			&result.ServiceID,
			&result.BoarID,
			&result.OperatorID,
			&result.MountNumber,
			&result.MountDate,
			&result.Type,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
		}
		mounts = append(mounts, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}
	return mounts, nil
}

// Create inserts the removal, deactivates the sow and fails its active service.
// All writes run in a single transaction so the rows always change together; the
// service may be nil when the sow has no confirmed service to fail.
func (r *PostgresSowRemovalRepository) Create(
	ctx context.Context,
	removal *sowremovaldomain.SowRemoval,
	sow *sowdomain.Sow,
	service *servicedomain.Service,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar la baja: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, `
		INSERT INTO sow_removals (
			id, sow_id, removal_date, type, reason, note, last_state,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		removal.ID,
		removal.SowID,
		removal.RemovalDate,
		removal.Type,
		removal.Reason,
		removal.Note,
		removal.LastState,
		removal.CreatedAt,
		removal.UpdatedAt,
		removal.CreatedBy,
		removal.UpdatedBy,
	); err != nil {
		return mapPostgresError(err)
	}

	commandTag, err := transaction.Exec(ctx, `
		UPDATE sows
		SET state = $2,
		    active = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		sow.ID,
		sow.State,
		sow.Active,
		sow.UpdatedAt,
		sow.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrSowNotFound
	}

	if service != nil {
		commandTag, err = transaction.Exec(ctx, `
			UPDATE services
			SET state = $2,
			    updated_at = $3,
			    updated_by = $4
			WHERE id = $1
		`,
			service.ID,
			service.State,
			service.UpdatedAt,
			service.UpdatedBy,
		)
		if err != nil {
			return mapPostgresError(err)
		}
		if commandTag.RowsAffected() == 0 {
			return ports.ErrServiceNotFound
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar la baja: %w", err)
	}
	return nil
}

// List fetches the removals matching the filter, newest first.
// It returns an empty slice when no removal matches and wraps any query failure.
func (r *PostgresSowRemovalRepository) List(ctx context.Context, filter ports.SowRemovalFilter) ([]*sowremovaldomain.SowRemoval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, sow_id, removal_date, type, reason, note, last_state,
		       created_at, updated_at, created_by, updated_by
		FROM sow_removals
		WHERE ($1::uuid IS NULL OR sow_id = $1)
		ORDER BY removal_date DESC, created_at DESC
	`, filter.SowID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las bajas: %w", err)
	}
	defer rows.Close()

	removals := make([]*sowremovaldomain.SowRemoval, 0)
	for rows.Next() {
		result := &sowremovaldomain.SowRemoval{}
		if err := rows.Scan(
			&result.ID,
			&result.SowID,
			&result.RemovalDate,
			&result.Type,
			&result.Reason,
			&result.Note,
			&result.LastState,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar las bajas: %w", err)
		}
		removals = append(removals, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las bajas: %w", err)
	}
	return removals, nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// A unique violation means the sow already has a removal; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrSowAlreadyRemoved
	}
	return fmt.Errorf("No se pudo guardar la baja: %w", err)
}
