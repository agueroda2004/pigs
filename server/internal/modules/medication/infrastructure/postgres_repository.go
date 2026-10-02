package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	medicationdomain "server/internal/modules/medication/domain"
	"server/internal/modules/medication/ports"
)

var ErrMedicationNotFound = ports.ErrMedicationNotFound

type PostgresMedicationRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresMedicationRepository creates a repository backed by a pgx pool.
// It stores the pool used for all medication queries.
func NewPostgresMedicationRepository(pool *pgxpool.Pool) *PostgresMedicationRepository {
	return &PostgresMedicationRepository{pool: pool}
}

// Create inserts a new medication row into the database.
// It maps unique constraint violations to ErrMedicationNameAlreadyUsed.
func (r *PostgresMedicationRepository) Create(ctx context.Context, medication *medicationdomain.Medication) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO medications (
			id, name, active, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		medication.ID,
		medication.Name,
		medication.Active,
		medication.CreatedAt,
		medication.UpdatedAt,
		medication.CreatedBy,
		medication.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	return nil
}

// GetByID fetches a single medication by its identifier.
// It returns ErrMedicationNotFound when no row matches.
func (r *PostgresMedicationRepository) GetByID(ctx context.Context, id uuid.UUID) (*medicationdomain.Medication, error) {
	result := &medicationdomain.Medication{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, active, created_at, updated_at, created_by, updated_by
		FROM medications
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
		return nil, ErrMedicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el medicamento: %w", err)
	}
	return result, nil
}

// ExistsByName reports whether a medication name is already taken.
// It wraps any query failure with context.
func (r *PostgresMedicationRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM medications WHERE name = $1
		)
	`, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("No se pudo verificar el nombre del medicamento: %w", err)
	}
	return exists, nil
}

// List fetches the medications matching the filter, ordered by name.
// It returns an empty slice when no medication matches and wraps any query failure.
func (r *PostgresMedicationRepository) List(ctx context.Context, filter ports.MedicationFilter) ([]*medicationdomain.Medication, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, active, created_at, updated_at, created_by, updated_by
		FROM medications
		WHERE ($1::text IS NULL OR name ILIKE '%' || $1 || '%')
		  AND ($2::boolean IS NULL OR active = $2)
		ORDER BY name ASC
	`,
		filter.Name,
		filter.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar los medicamentos: %w", err)
	}
	defer rows.Close()

	medications := make([]*medicationdomain.Medication, 0)
	for rows.Next() {
		result := &medicationdomain.Medication{}
		if err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Active,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudo consultar los medicamentos: %w", err)
		}
		medications = append(medications, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudo consultar los medicamentos: %w", err)
	}
	return medications, nil
}

// ListOptions fetches the id and name of the medications matching the active filter.
// A true active restricts the result to active medications; nil applies no filter
// and returns every medication, active or inactive.
func (r *PostgresMedicationRepository) ListOptions(ctx context.Context, active *bool) ([]medicationdomain.MedicationOption, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name
		FROM medications
		WHERE ($1::boolean IS NOT TRUE OR active = TRUE)
		ORDER BY name ASC
	`, active)
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar los medicamentos: %w", err)
	}
	defer rows.Close()

	options := make([]medicationdomain.MedicationOption, 0)
	for rows.Next() {
		var option medicationdomain.MedicationOption
		if err := rows.Scan(&option.ID, &option.Name); err != nil {
			return nil, fmt.Errorf("No se pudo consultar los medicamentos: %w", err)
		}
		options = append(options, option)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudo consultar los medicamentos: %w", err)
	}
	return options, nil
}

// Update persists the medication's mutable fields by id.
// It returns ErrMedicationNotFound when no row was affected.
func (r *PostgresMedicationRepository) Update(ctx context.Context, medication *medicationdomain.Medication) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE medications
		SET name = $2,
		    active = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		medication.ID,
		medication.Name,
		medication.Active,
		medication.UpdatedAt,
		medication.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrMedicationNotFound
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// Unique violations become ErrMedicationNameAlreadyUsed; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrMedicationNameAlreadyUsed
	}
	return fmt.Errorf("No se pudo guardar el medicamento: %w", err)
}
