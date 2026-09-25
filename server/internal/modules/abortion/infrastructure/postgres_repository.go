package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	abortiondomain "server/internal/modules/abortion/domain"
	"server/internal/modules/abortion/ports"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

var ErrAbortionNotFound = ports.ErrAbortionNotFound

type PostgresAbortionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAbortionRepository creates a repository backed by a pgx pool.
// It stores the pool used for all abortion, sow and service queries.
func NewPostgresAbortionRepository(pool *pgxpool.Pool) *PostgresAbortionRepository {
	return &PostgresAbortionRepository{pool: pool}
}

// GetSow fetches the sow referenced by an abortion by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresAbortionRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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
func (r *PostgresAbortionRepository) GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error) {
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

// listMounts fetches every mount of a service ordered by mount number.
// It returns an empty slice when the service has no mounts.
func (r *PostgresAbortionRepository) listMounts(ctx context.Context, serviceID uuid.UUID) ([]*servicedomain.Mount, error) {
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

// Create inserts the abortion and moves the sow and its service to aborted.
// All writes run in a single transaction so the three rows always change together.
func (r *PostgresAbortionRepository) Create(
	ctx context.Context,
	abortion *abortiondomain.Abortion,
	sow *sowdomain.Sow,
	service *servicedomain.Service,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el aborto: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, `
		INSERT INTO abortions (
			id, sow_id, service_id, abortion_date, cause, note,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		abortion.ID,
		abortion.SowID,
		abortion.ServiceID,
		abortion.AbortionDate,
		abortion.Cause,
		abortion.Note,
		abortion.CreatedAt,
		abortion.UpdatedAt,
		abortion.CreatedBy,
		abortion.UpdatedBy,
	); err != nil {
		return mapPostgresError(err)
	}

	commandTag, err := transaction.Exec(ctx, `
		UPDATE sows
		SET state = $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`,
		sow.ID,
		sow.State,
		sow.UpdatedAt,
		sow.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrSowNotFound
	}

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

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el aborto: %w", err)
	}
	return nil
}

// List fetches the abortions matching the filter, newest first.
// It returns an empty slice when no abortion matches and wraps any query failure.
func (r *PostgresAbortionRepository) List(ctx context.Context, filter ports.AbortionFilter) ([]*abortiondomain.Abortion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, sow_id, service_id, abortion_date, cause, note,
		       created_at, updated_at, created_by, updated_by
		FROM abortions
		WHERE ($1::uuid IS NULL OR sow_id = $1)
		ORDER BY abortion_date DESC, created_at DESC
	`, filter.SowID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los abortos: %w", err)
	}
	defer rows.Close()

	abortions := make([]*abortiondomain.Abortion, 0)
	for rows.Next() {
		result := &abortiondomain.Abortion{}
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los abortos: %w", err)
		}
		abortions = append(abortions, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los abortos: %w", err)
	}
	return abortions, nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// It wraps any error with context because no module-specific constraint exists.
func mapPostgresError(err error) error {
	return fmt.Errorf("No se pudo guardar el aborto: %w", err)
}
