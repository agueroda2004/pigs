package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	boardomain "server/internal/modules/boar/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	"server/internal/modules/service/ports"
	sowdomain "server/internal/modules/sow/domain"
)

var (
	ErrServiceNotFound = ports.ErrServiceNotFound
)

type PostgresServiceRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresServiceRepository creates a repository backed by a pgx pool.
// It stores the pool used for all service and mount queries.
func NewPostgresServiceRepository(pool *pgxpool.Pool) *PostgresServiceRepository {
	return &PostgresServiceRepository{pool: pool}
}

// GetSow fetches the sow referenced by a service by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresServiceRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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

// GetBoar fetches the boar referenced by a mount by its identifier.
// It returns ports.ErrBoarNotFound when no row matches.
func (r *PostgresServiceRepository) GetBoar(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error) {
	result := &boardomain.Boar{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, code, location, active, entry_date, birth_date, note, state,
		       origin, breed_id, created_at, updated_at, created_by, updated_by
		FROM boars
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
		&result.BreedID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrBoarNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el verraco: %w", err)
	}
	return result, nil
}

// GetOperator fetches the operator referenced by a mount by its identifier.
// It returns ports.ErrOperatorNotFound when no row matches.
func (r *PostgresServiceRepository) GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error) {
	result := &operatordomain.Operator{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, active, created_at, updated_at, created_by, updated_by
		FROM operators
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&result.Name,
		&result.Active,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrOperatorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el operador: %w", err)
	}
	return result, nil
}

// GetByID fetches a single service with its mounts by its identifier.
// It returns ports.ErrServiceNotFound when no row matches.
func (r *PostgresServiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*servicedomain.Service, error) {
	result, err := scanService(r.pool.QueryRow(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location, last_state,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE id = $1
	`, id))
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

// GetLastService fetches the most recent service of a sow with its mounts.
// It returns ports.ErrServiceNotFound when the sow has no service.
func (r *PostgresServiceRepository) GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error) {
	return r.serviceWithMounts(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location, last_state,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE sow_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, sowID)
}

// GetPreviousService fetches the most recent service of a sow other than the
// excluded one, with its mounts. It returns ports.ErrServiceNotFound when none exists.
func (r *PostgresServiceRepository) GetPreviousService(ctx context.Context, sowID uuid.UUID, excludeID uuid.UUID) (*servicedomain.Service, error) {
	return r.serviceWithMounts(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location, last_state,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE sow_id = $1 AND id <> $2
		ORDER BY created_at DESC
		LIMIT 1
	`, sowID, excludeID)
}

// GetLastAbortionDate returns the most recent abortion date of a sow.
// It returns nil when the sow has no abortion.
func (r *PostgresServiceRepository) GetLastAbortionDate(ctx context.Context, sowID uuid.UUID) (*time.Time, error) {
	var lastAbortion *time.Time
	if err := r.pool.QueryRow(ctx, `
		SELECT MAX(abortion_date)
		FROM abortions
		WHERE sow_id = $1
	`, sowID).Scan(&lastAbortion); err != nil {
		return nil, fmt.Errorf("No se pudo consultar el último aborto: %w", err)
	}
	return lastAbortion, nil
}

// serviceWithMounts runs a single-row service query and loads its mounts.
// It returns ports.ErrServiceNotFound when no row matches.
func (r *PostgresServiceRepository) serviceWithMounts(ctx context.Context, query string, args ...any) (*servicedomain.Service, error) {
	result, err := scanService(r.pool.QueryRow(ctx, query, args...))
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

// scanService scans a service row without its mounts.
// It returns the pgx no-rows error when the query has no result.
func scanService(row pgx.Row) (*servicedomain.Service, error) {
	result := &servicedomain.Service{}
	err := row.Scan(
		&result.ID,
		&result.SowID,
		&result.ExpectedFarrowingDate,
		&result.Note,
		&result.State,
		&result.Location,
		&result.LastState,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// listMounts fetches every mount of a service ordered by mount number.
// It returns an empty slice when the service has no mounts.
func (r *PostgresServiceRepository) listMounts(ctx context.Context, serviceID uuid.UUID) ([]*servicedomain.Mount, error) {
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

// Create inserts the service and its mounts and moves the sow to its new state.
// All writes run in a single transaction so a service never exists without its
// mounts or the matching sow state change.
func (r *PostgresServiceRepository) Create(ctx context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el servicio: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, `
		INSERT INTO services (
			id, sow_id, expected_farrowing_date, note, state, location, last_state,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		service.ID,
		service.SowID,
		service.ExpectedFarrowingDate,
		service.Note,
		service.State,
		service.Location,
		service.LastState,
		service.CreatedAt,
		service.UpdatedAt,
		service.CreatedBy,
		service.UpdatedBy,
	); err != nil {
		return mapPostgresError(err)
	}

	for _, mount := range service.Mounts {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO mounts (
				id, service_id, boar_id, operator_id, mount_number, mount_date,
				type, note, created_at, updated_at, created_by, updated_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
			mount.ID,
			mount.ServiceID,
			mount.BoarID,
			mount.OperatorID,
			mount.MountNumber,
			mount.MountDate,
			mount.Type,
			mount.Note,
			mount.CreatedAt,
			mount.UpdatedAt,
			mount.CreatedBy,
			mount.UpdatedBy,
		); err != nil {
			return mapPostgresError(err)
		}
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

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el servicio: %w", err)
	}
	return nil
}

// List fetches one page of services matching the filter together with their mounts
// and the total number of matches. It returns an empty slice when no service matches.
func (r *PostgresServiceRepository) List(ctx context.Context, filter ports.ServiceFilter, limit, offset int) ([]*servicedomain.Service, int, error) {
	var state *string
	if filter.State != nil {
		value := string(*filter.State)
		state = &value
	}

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM services sv
		JOIN sows sow ON sow.id = sv.sow_id
		WHERE ($1::text IS NULL OR LOWER(sow.code) = LOWER($1))
		  AND ($2::text IS NULL OR sv.state::text = $2)
	`,
		filter.SowCode,
		state,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT sv.id, sv.sow_id, sv.expected_farrowing_date, sv.note, sv.state,
		       sv.location, sv.last_state, sv.created_at, sv.updated_at,
		       sv.created_by, sv.updated_by, sow.code
		FROM services sv
		JOIN sows sow ON sow.id = sv.sow_id
		WHERE ($1::text IS NULL OR LOWER(sow.code) = LOWER($1))
		  AND ($2::text IS NULL OR sv.state::text = $2)
		ORDER BY sv.created_at DESC
		LIMIT $3 OFFSET $4
	`,
		filter.SowCode,
		state,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
	}
	defer rows.Close()

	services := make([]*servicedomain.Service, 0)
	index := make(map[uuid.UUID]*servicedomain.Service)
	ids := make([]uuid.UUID, 0)

	for rows.Next() {
		result := &servicedomain.Service{}
		if err := rows.Scan(
			&result.ID,
			&result.SowID,
			&result.ExpectedFarrowingDate,
			&result.Note,
			&result.State,
			&result.Location,
			&result.LastState,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
			&result.SowCode,
		); err != nil {
			return nil, 0, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
		}
		result.Mounts = make([]*servicedomain.Mount, 0)
		services = append(services, result)
		index[result.ID] = result
		ids = append(ids, result.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
	}

	if len(ids) == 0 {
		return services, total, nil
	}

	mountRows, err := r.pool.Query(ctx, `
		SELECT m.id, m.service_id, m.boar_id, m.operator_id, m.mount_number, m.mount_date,
		       m.type, m.note, m.created_at, m.updated_at, m.created_by, m.updated_by,
		       b.code, o.name
		FROM mounts m
		JOIN boars b ON b.id = m.boar_id
		JOIN operators o ON o.id = m.operator_id
		WHERE m.service_id = ANY($1)
		ORDER BY m.service_id, m.mount_number ASC
	`, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}
	defer mountRows.Close()

	for mountRows.Next() {
		result := &servicedomain.Mount{}
		if err := mountRows.Scan(
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
			&result.BoarCode,
			&result.OperatorName,
		); err != nil {
			return nil, 0, fmt.Errorf("No se pudieron consultar las montas: %w", err)
		}
		if service, ok := index[result.ServiceID]; ok {
			service.Mounts = append(service.Mounts, result)
		}
	}
	if err := mountRows.Err(); err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}

	return services, total, nil
}

// Update persists the service fields and replaces its mounts.
// All writes run in a single transaction: the service row is updated and every
// mount is reinserted from the service's final list, which supports create,
// update and delete operations while preserving the kept mounts' audit fields.
func (r *PostgresServiceRepository) Update(ctx context.Context, service *servicedomain.Service) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo actualizar el servicio: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE services
		SET location = $2,
		    note = $3,
		    expected_farrowing_date = $4,
		    updated_at = $5,
		    updated_by = $6
		WHERE id = $1
	`,
		service.ID,
		service.Location,
		service.Note,
		service.ExpectedFarrowingDate,
		service.UpdatedAt,
		service.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrServiceNotFound
	}

	if _, err := transaction.Exec(ctx, `
		DELETE FROM mounts
		WHERE service_id = $1
	`, service.ID); err != nil {
		return mapPostgresError(err)
	}

	for _, mount := range service.Mounts {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO mounts (
				id, service_id, boar_id, operator_id, mount_number, mount_date,
				type, note, created_at, updated_at, created_by, updated_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
			mount.ID,
			mount.ServiceID,
			mount.BoarID,
			mount.OperatorID,
			mount.MountNumber,
			mount.MountDate,
			mount.Type,
			mount.Note,
			mount.CreatedAt,
			mount.UpdatedAt,
			mount.CreatedBy,
			mount.UpdatedBy,
		); err != nil {
			return mapPostgresError(err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo actualizar el servicio: %w", err)
	}
	return nil
}

// Delete removes a service and its mounts and restores the sow state.
// All writes run in a single transaction; a foreign key violation on the service
// means it still has related records and maps to ports.ErrServiceInUse.
func (r *PostgresServiceRepository) Delete(ctx context.Context, service *servicedomain.Service, sow *sowdomain.Sow) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar el servicio: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

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

	if _, err := transaction.Exec(ctx, `
		DELETE FROM mounts
		WHERE service_id = $1
	`, service.ID); err != nil {
		return mapPostgresError(err)
	}

	commandTag, err = transaction.Exec(ctx, `
		DELETE FROM services
		WHERE id = $1
	`, service.ID)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23503" {
			return ports.ErrServiceInUse
		}
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrServiceNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo eliminar el servicio: %w", err)
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// It wraps any error with context because no module-specific constraint exists.
func mapPostgresError(err error) error {
	return fmt.Errorf("No se pudo guardar el servicio: %w", err)
}
