package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	boardomain "server/internal/modules/boar/domain"
	"server/internal/modules/boar/ports"
)

var ErrBoarNotFound = ports.ErrBoarNotFound

type PostgresBoarRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresBoarRepository creates a repository backed by a pgx pool.
// It stores the pool used for all boar queries.
func NewPostgresBoarRepository(pool *pgxpool.Pool) *PostgresBoarRepository {
	return &PostgresBoarRepository{pool: pool}
}

// Create inserts a new boar row into the database.
// It maps unique constraint violations to ErrBoarCodeAlreadyUsed.
func (r *PostgresBoarRepository) Create(ctx context.Context, boar *boardomain.Boar) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO boars (
			id, code, location, active, entry_date, birth_date, note, state,
			origin, breed_id, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		boar.ID,
		boar.Code,
		boar.Location,
		boar.Active,
		boar.EntryDate,
		boar.BirthDate,
		boar.Note,
		boar.State,
		boar.Origin,
		boar.BreedID,
		boar.CreatedAt,
		boar.UpdatedAt,
		boar.CreatedBy,
		boar.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	return nil
}

// GetByID fetches a single boar by its identifier.
// It returns ErrBoarNotFound when no row matches.
func (r *PostgresBoarRepository) GetByID(ctx context.Context, id uuid.UUID) (*boardomain.Boar, error) {
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
		return nil, ErrBoarNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el verraco: %w", err)
	}
	return result, nil
}

// ExistsByCode reports whether a boar code is already taken.
// It wraps any query failure with context.
func (r *PostgresBoarRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM boars WHERE code = $1
		)
	`, code).Scan(&exists); err != nil {
		return false, fmt.Errorf("No se pudo verificar el código del verraco: %w", err)
	}
	return exists, nil
}

// List fetches the boars matching the filter, ordered by code.
// It returns an empty slice when no boar matches and wraps any query failure.
func (r *PostgresBoarRepository) List(ctx context.Context, filter ports.BoarFilter) ([]*boardomain.Boar, error) {
	var origin *string
	if filter.Origin != nil {
		value := string(*filter.Origin)
		origin = &value
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, code, location, active, entry_date, birth_date, note, state,
		       origin, breed_id, created_at, updated_at, created_by, updated_by
		FROM boars
		WHERE ($1::text IS NULL OR code ILIKE '%' || $1 || '%')
		  AND ($2::uuid IS NULL OR breed_id = $2)
		  AND ($3::text IS NULL OR origin::text = $3)
		  AND ($4::boolean IS NULL OR active = $4)
		ORDER BY code ASC
	`,
		filter.Code,
		filter.BreedID,
		origin,
		filter.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar los verracos: %w", err)
	}
	defer rows.Close()

	boars := make([]*boardomain.Boar, 0)
	for rows.Next() {
		result := &boardomain.Boar{}
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
			&result.BreedID,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudo consultar los verracos: %w", err)
		}
		boars = append(boars, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudo consultar los verracos: %w", err)
	}
	return boars, nil
}

// Update persists the boar's user-mutable fields by id, never its state.
// It returns ErrBoarNotFound when no row was affected.
func (r *PostgresBoarRepository) Update(ctx context.Context, boar *boardomain.Boar) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE boars
		SET code = $2,
		    location = $3,
		    active = $4,
		    entry_date = $5,
		    birth_date = $6,
		    note = $7,
		    origin = $8,
		    breed_id = $9,
		    updated_at = $10,
		    updated_by = $11
		WHERE id = $1
	`,
		boar.ID,
		boar.Code,
		boar.Location,
		boar.Active,
		boar.EntryDate,
		boar.BirthDate,
		boar.Note,
		boar.Origin,
		boar.BreedID,
		boar.UpdatedAt,
		boar.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrBoarNotFound
	}
	return nil
}

// UpdateState persists only the boar state and its audit fields by id.
// It is the single write path for state and returns ErrBoarNotFound when missing.
func (r *PostgresBoarRepository) UpdateState(ctx context.Context, boar *boardomain.Boar) error {
	commandTag, err := r.pool.Exec(ctx, `
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
		return ErrBoarNotFound
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// Unique violations become ErrBoarCodeAlreadyUsed; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrBoarCodeAlreadyUsed
	}
	return fmt.Errorf("No se pudo guardar el verraco: %w", err)
}
