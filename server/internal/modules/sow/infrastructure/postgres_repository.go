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

	sowdomain "server/internal/modules/sow/domain"
	"server/internal/modules/sow/ports"
)

var ErrSowNotFound = ports.ErrSowNotFound

type PostgresSowRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSowRepository creates a repository backed by a pgx pool.
// It stores the pool used for all sow queries.
func NewPostgresSowRepository(pool *pgxpool.Pool) *PostgresSowRepository {
	return &PostgresSowRepository{pool: pool}
}

// Create inserts a new sow row into the database.
// It maps unique constraint violations to ErrSowCodeAlreadyUsed.
func (r *PostgresSowRepository) Create(ctx context.Context, sow *sowdomain.Sow) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sows (
			id, code, location, active, entry_date, birth_date, note, state,
			origin, parity, breed_id, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`,
		sow.ID,
		sow.Code,
		sow.Location,
		sow.Active,
		sow.EntryDate,
		sow.BirthDate,
		sow.Note,
		sow.State,
		sow.Origin,
		sow.Parity,
		sow.BreedID,
		sow.CreatedAt,
		sow.UpdatedAt,
		sow.CreatedBy,
		sow.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	return nil
}

// GetByID fetches a single sow by its identifier.
// It returns ErrSowNotFound when no row matches.
func (r *PostgresSowRepository) GetByID(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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
		return nil, ErrSowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar la cerda: %w", err)
	}
	return result, nil
}

// ExistsByCode reports whether a sow code is already taken.
// It wraps any query failure with context.
func (r *PostgresSowRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sows WHERE code = $1
		)
	`, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("No se pudo verificar el código de la cerda: %w", err)
	}
	return exists, nil
}

// List fetches one page of sows matching the filter, ordered by code, plus the
// total number of matches. It returns an empty slice when no sow matches.
func (r *PostgresSowRepository) List(ctx context.Context, filter ports.SowFilter, limit, offset int) ([]*sowdomain.Sow, int, error) {
	var origin *string
	if filter.Origin != nil {
		value := string(*filter.Origin)
		origin = &value
	}

	var state *string
	if filter.State != nil {
		value := string(*filter.State)
		state = &value
	}

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM sows
		WHERE ($1::text IS NULL OR code ILIKE '%' || $1 || '%')
		  AND ($2::uuid IS NULL OR breed_id = $2)
		  AND ($3::text IS NULL OR origin::text = $3)
		  AND ($4::boolean IS NULL OR active = $4)
		  AND ($5::text IS NULL OR state::text = $5)
	`,
		filter.Code,
		filter.BreedID,
		origin,
		filter.Active,
		state,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, code, location, active, entry_date, birth_date, note, state,
		       origin, parity, breed_id, created_at, updated_at, created_by, updated_by
		FROM sows
		WHERE ($1::text IS NULL OR code ILIKE '%' || $1 || '%')
		  AND ($2::uuid IS NULL OR breed_id = $2)
		  AND ($3::text IS NULL OR origin::text = $3)
		  AND ($4::boolean IS NULL OR active = $4)
		  AND ($5::text IS NULL OR state::text = $5)
		ORDER BY code ASC
		LIMIT $6 OFFSET $7
	`,
		filter.Code,
		filter.BreedID,
		origin,
		filter.Active,
		state,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
	}
	defer rows.Close()

	sows := make([]*sowdomain.Sow, 0)
	for rows.Next() {
		result := &sowdomain.Sow{}
		if err := rows.Scan(
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
		); err != nil {
			return nil, 0, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
		}
		sows = append(sows, result)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
	}
	return sows, total, nil
}

// ListDropdown fetches the id and code of the sows matching the active flag and
// state list. A nil active and an empty state list apply no filter.
func (r *PostgresSowRepository) ListDropdown(ctx context.Context, active *bool, states []sowdomain.State) ([]sowdomain.SowDropdown, error) {
	query := `
		SELECT id, code
		FROM sows
		WHERE ($1::boolean IS NULL OR active = $1)
		ORDER BY code ASC
	`
	args := []any{active}

	if len(states) > 0 {
		stateValues := make([]string, 0, len(states))
		for _, state := range states {
			stateValues = append(stateValues, string(state))
		}
		query = `
		SELECT id, code
		FROM sows
		WHERE ($1::boolean IS NULL OR active = $1)
		  AND state::text = ANY($2)
		ORDER BY code ASC
	`
		args = append(args, stateValues)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
	}
	defer rows.Close()

	options := make([]sowdomain.SowDropdown, 0)
	for rows.Next() {
		var option sowdomain.SowDropdown
		if err := rows.Scan(&option.ID, &option.Code); err != nil {
			return nil, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
		}
		options = append(options, option)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudo consultar las cerdas: %w", err)
	}
	return options, nil
}

// LastServiceDate fetches the most recent mount date registered for a sow.
// It returns nil when the sow has no service or mount yet.
func (r *PostgresSowRepository) LastServiceDate(ctx context.Context, sowID uuid.UUID) (*time.Time, error) {
	var lastServiceDate *time.Time
	if err := r.pool.QueryRow(ctx, `
		SELECT MAX(mounts.mount_date)
		FROM mounts
		JOIN services ON services.id = mounts.service_id
		WHERE services.sow_id = $1
	`, sowID).Scan(&lastServiceDate); err != nil {
		return nil, fmt.Errorf("No se pudo consultar el último servicio: %w", err)
	}
	return lastServiceDate, nil
}

// Update persists the sow's user-mutable fields by id, never its active flag or state.
// It returns ErrSowNotFound when no row was affected.
func (r *PostgresSowRepository) Update(ctx context.Context, sow *sowdomain.Sow) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE sows
		SET code = $2,
		    location = $3,
		    entry_date = $4,
		    birth_date = $5,
		    note = $6,
		    origin = $7,
		    breed_id = $8,
		    parity = $9,
		    updated_at = $10,
		    updated_by = $11
		WHERE id = $1
	`,
		sow.ID,
		sow.Code,
		sow.Location,
		sow.EntryDate,
		sow.BirthDate,
		sow.Note,
		sow.Origin,
		sow.BreedID,
		sow.Parity,
		sow.UpdatedAt,
		sow.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrSowNotFound
	}
	return nil
}

// UpdateState persists only the sow state and its audit fields by id.
// It is the single write path for state and returns ErrSowNotFound when missing.
func (r *PostgresSowRepository) UpdateState(ctx context.Context, sow *sowdomain.Sow) error {
	commandTag, err := r.pool.Exec(ctx, `
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
		return ErrSowNotFound
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// Unique violations become ErrSowCodeAlreadyUsed; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrSowCodeAlreadyUsed
	}
	return fmt.Errorf("No se pudo guardar la cerda: %w", err)
}
