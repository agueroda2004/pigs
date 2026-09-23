package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	breeddomain "server/internal/modules/breed/domain"
	"server/internal/modules/breed/ports"
)

var ErrBreedNotFound = ports.ErrBreedNotFound

type PostgresBreedRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresBreedRepository creates a repository backed by a pgx pool.
// It stores the pool used for all breed queries.
func NewPostgresBreedRepository(pool *pgxpool.Pool) *PostgresBreedRepository {
	return &PostgresBreedRepository{pool: pool}
}

// Create inserts a new breed row into the database.
// It maps unique constraint violations to ErrBreedNameAlreadyUsed.
func (r *PostgresBreedRepository) Create(ctx context.Context, breed *breeddomain.Breed) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO breeds (
			id, name, active, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		breed.ID,
		breed.Name,
		breed.Active,
		breed.CreatedAt,
		breed.UpdatedAt,
		breed.CreatedBy,
		breed.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	return nil
}

// GetByID fetches a single breed by its identifier.
// It returns ErrBreedNotFound when no row matches.
func (r *PostgresBreedRepository) GetByID(ctx context.Context, id uuid.UUID) (*breeddomain.Breed, error) {
	result := &breeddomain.Breed{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, active, created_at, updated_at, created_by, updated_by
		FROM breeds
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
		return nil, ErrBreedNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar la raza: %w", err)
	}
	return result, nil
}

// ExistsByName reports whether a breed name is already taken.
// It wraps any query failure with context.
func (r *PostgresBreedRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM breeds WHERE name = $1
		)
	`, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("No se pudo verificar el nombre de la raza: %w", err)
	}
	return exists, nil
}

// List fetches every breed ordered by name.
// It returns an empty slice when no breeds exist and wraps any query failure.
func (r *PostgresBreedRepository) List(ctx context.Context) ([]*breeddomain.Breed, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, active, created_at, updated_at, created_by, updated_by
		FROM breeds
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar las razas: %w", err)
	}
	defer rows.Close()

	breeds := make([]*breeddomain.Breed, 0)
	for rows.Next() {
		result := &breeddomain.Breed{}
		if err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Active,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudo consultar las razas: %w", err)
		}
		breeds = append(breeds, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudo consultar las razas: %w", err)
	}
	return breeds, nil
}

// Update persists the breed's mutable fields by id.
// It returns ErrBreedNotFound when no row was affected.
func (r *PostgresBreedRepository) Update(ctx context.Context, breed *breeddomain.Breed) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE breeds
		SET name = $2,
		    active = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		breed.ID,
		breed.Name,
		breed.Active,
		breed.UpdatedAt,
		breed.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrBreedNotFound
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// Unique violations become ErrBreedNameAlreadyUsed; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrBreedNameAlreadyUsed
	}
	return fmt.Errorf("No se pudo guardar la raza: %w", err)
}
