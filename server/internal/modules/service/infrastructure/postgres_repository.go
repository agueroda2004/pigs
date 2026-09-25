package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
			id, sow_id, expected_farrowing_date, note, state, location,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		service.ID,
		service.SowID,
		service.ExpectedFarrowingDate,
		service.Note,
		service.State,
		service.Location,
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

// List fetches the services matching the filter together with their mounts.
// It returns an empty slice when no service matches and wraps any query failure.
func (r *PostgresServiceRepository) List(ctx context.Context, filter ports.ServiceFilter) ([]*servicedomain.Service, error) {
	var state *string
	if filter.State != nil {
		value := string(*filter.State)
		state = &value
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE ($1::uuid IS NULL OR sow_id = $1)
		  AND ($2::text IS NULL OR state::text = $2)
		ORDER BY created_at DESC
	`,
		filter.SowID,
		state,
	)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
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
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
		}
		result.Mounts = make([]*servicedomain.Mount, 0)
		services = append(services, result)
		index[result.ID] = result
		ids = append(ids, result.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los servicios: %w", err)
	}

	if len(ids) == 0 {
		return services, nil
	}

	mountRows, err := r.pool.Query(ctx, `
		SELECT id, service_id, boar_id, operator_id, mount_number, mount_date,
		       type, note, created_at, updated_at, created_by, updated_by
		FROM mounts
		WHERE service_id = ANY($1)
		ORDER BY service_id, mount_number ASC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
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
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
		}
		if service, ok := index[result.ServiceID]; ok {
			service.Mounts = append(service.Mounts, result)
		}
	}
	if err := mountRows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}

	return services, nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// It wraps any error with context because no module-specific constraint exists.
func mapPostgresError(err error) error {
	return fmt.Errorf("No se pudo guardar el servicio: %w", err)
}
