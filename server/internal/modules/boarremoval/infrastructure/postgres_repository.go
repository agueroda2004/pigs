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
	boarremovaldomain "server/internal/modules/boarremoval/domain"
	"server/internal/modules/boarremoval/ports"
)

var ErrBoarRemovalNotFound = ports.ErrBoarRemovalNotFound

type PostgresBoarRemovalRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresBoarRemovalRepository creates a repository backed by a pgx pool.
// It stores the pool used for all removal, boar and mount queries.
func NewPostgresBoarRemovalRepository(pool *pgxpool.Pool) *PostgresBoarRemovalRepository {
	return &PostgresBoarRemovalRepository{pool: pool}
}

// GetByID fetches the removal referenced by its identifier.
// It returns ports.ErrBoarRemovalNotFound when no row matches.
func (r *PostgresBoarRemovalRepository) GetByID(ctx context.Context, id uuid.UUID) (*boarremovaldomain.BoarRemoval, error) {
	result := &boarremovaldomain.BoarRemoval{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, boar_id, removal_date, type, reason, note, last_state,
		       created_at, updated_at, created_by, updated_by
		FROM boar_removals
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&result.BoarID,
		&result.RemovalDate,
		&result.Type,
		&result.Reason,
		&result.Note,
		&result.LastState,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrBoarRemovalNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar la baja: %w", err)
	}
	return result, nil
}

// GetBoar fetches the boar referenced by a removal by its identifier.
// It returns ports.ErrBoarNotFound when no row matches.
func (r *PostgresBoarRemovalRepository) GetBoar(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error) {
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

// GetLastMountDate returns the latest mount date of a boar across all services.
// It returns the zero time when the boar has no mounts and wraps any query failure.
func (r *PostgresBoarRemovalRepository) GetLastMountDate(ctx context.Context, boarID uuid.UUID) (time.Time, error) {
	var lastMount *time.Time

	if err := r.pool.QueryRow(ctx, `
		SELECT MAX(mount_date)
		FROM mounts
		WHERE boar_id = $1
	`, boarID).Scan(&lastMount); err != nil {
		return time.Time{}, fmt.Errorf("No se pudo consultar la última monta: %w", err)
	}
	if lastMount == nil {
		return time.Time{}, nil
	}
	return *lastMount, nil
}

// Create inserts the removal and deactivates the boar in a single transaction.
// It maps a duplicate removal to ErrBoarAlreadyRemoved and a missing boar to
// ErrBoarNotFound so the two rows always change together.
func (r *PostgresBoarRemovalRepository) Create(
	ctx context.Context,
	removal *boarremovaldomain.BoarRemoval,
	boar *boardomain.Boar,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar la baja: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, `
		INSERT INTO boar_removals (
			id, boar_id, removal_date, type, reason, note, last_state,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		removal.ID,
		removal.BoarID,
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
		UPDATE boars
		SET state = $2,
		    active = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		boar.ID,
		boar.State,
		boar.Active,
		boar.UpdatedAt,
		boar.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrBoarNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar la baja: %w", err)
	}
	return nil
}

// Update persists the removal fields and, when provided, the boar state change.
// Both writes run in a single transaction so the rows always change together;
// the boar may be nil when the removal type did not change.
func (r *PostgresBoarRemovalRepository) Update(
	ctx context.Context,
	removal *boarremovaldomain.BoarRemoval,
	boar *boardomain.Boar,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo actualizar la baja: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE boar_removals
		SET removal_date = $2,
		    type = $3,
		    reason = $4,
		    note = $5,
		    updated_at = $6,
		    updated_by = $7
		WHERE id = $1
	`,
		removal.ID,
		removal.RemovalDate,
		removal.Type,
		removal.Reason,
		removal.Note,
		removal.UpdatedAt,
		removal.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrBoarRemovalNotFound
	}

	if boar != nil {
		commandTag, err = transaction.Exec(ctx, `
			UPDATE boars
			SET state = $2,
			    updated_at = $3,
			    updated_by = $4
			WHERE id = $1
		`,
			boar.ID,
			boar.State,
			boar.UpdatedAt,
			boar.UpdatedBy,
		)
		if err != nil {
			return mapPostgresError(err)
		}
		if commandTag.RowsAffected() == 0 {
			return ports.ErrBoarNotFound
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo actualizar la baja: %w", err)
	}
	return nil
}

// Delete removes the removal and undoes its effects on the boar.
// Both writes run in a single transaction so the rows always change together.
func (r *PostgresBoarRemovalRepository) Delete(
	ctx context.Context,
	removal *boarremovaldomain.BoarRemoval,
	boar *boardomain.Boar,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar la baja: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		DELETE FROM boar_removals
		WHERE id = $1
	`, removal.ID)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrBoarRemovalNotFound
	}

	commandTag, err = transaction.Exec(ctx, `
		UPDATE boars
		SET state = $2,
		    active = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		boar.ID,
		boar.State,
		boar.Active,
		boar.UpdatedAt,
		boar.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrBoarNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo eliminar la baja: %w", err)
	}
	return nil
}

// List fetches the removals matching the filter, newest first.
// It returns an empty slice when no removal matches and wraps any query failure.
func (r *PostgresBoarRemovalRepository) List(ctx context.Context, filter ports.BoarRemovalFilter) ([]*boarremovaldomain.BoarRemoval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, boar_id, removal_date, type, reason, note, last_state,
		       created_at, updated_at, created_by, updated_by
		FROM boar_removals
		WHERE ($1::uuid IS NULL OR boar_id = $1)
		ORDER BY removal_date DESC, created_at DESC
	`, filter.BoarID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las bajas: %w", err)
	}
	defer rows.Close()

	removals := make([]*boarremovaldomain.BoarRemoval, 0)
	for rows.Next() {
		result := &boarremovaldomain.BoarRemoval{}
		if err := rows.Scan(
			&result.ID,
			&result.BoarID,
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
// A unique violation means the boar already has a removal; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrBoarAlreadyRemoved
	}
	return fmt.Errorf("No se pudo guardar la baja: %w", err)
}
