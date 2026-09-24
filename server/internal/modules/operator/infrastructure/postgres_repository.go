package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	operatordomain "server/internal/modules/operator/domain"
	"server/internal/modules/operator/ports"
)

var ErrOperatorNotFound = ports.ErrOperatorNotFound

type PostgresOperatorRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresOperatorRepository creates a repository backed by a pgx pool.
// It stores the pool used for all operator queries.
func NewPostgresOperatorRepository(pool *pgxpool.Pool) *PostgresOperatorRepository {
	return &PostgresOperatorRepository{pool: pool}
}

// Create inserts a new operator row into the database.
// It maps unique constraint violations to ErrOperatorNameAlreadyUsed.
func (r *PostgresOperatorRepository) Create(ctx context.Context, operator *operatordomain.Operator) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO operators (
			id, name, active, created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		operator.ID,
		operator.Name,
		operator.Active,
		operator.CreatedAt,
		operator.UpdatedAt,
		operator.CreatedBy,
		operator.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	return nil
}

// GetByID fetches a single operator by its identifier.
// It returns ErrOperatorNotFound when no row matches.
func (r *PostgresOperatorRepository) GetByID(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error) {
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
		return nil, ErrOperatorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el operador: %w", err)
	}
	return result, nil
}

// ExistsByName reports whether an operator name is already taken.
// It wraps any query failure with context.
func (r *PostgresOperatorRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM operators WHERE name = $1
		)
	`, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("No se pudo verificar el nombre del operador: %w", err)
	}
	return exists, nil
}

// List fetches every operator ordered by name.
// It returns an empty slice when no operators exist and wraps any query failure.
func (r *PostgresOperatorRepository) List(ctx context.Context) ([]*operatordomain.Operator, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, active, created_at, updated_at, created_by, updated_by
		FROM operators
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar los operadores: %w", err)
	}
	defer rows.Close()

	operators := make([]*operatordomain.Operator, 0)
	for rows.Next() {
		result := &operatordomain.Operator{}
		if err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Active,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudo consultar los operadores: %w", err)
		}
		operators = append(operators, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudo consultar los operadores: %w", err)
	}
	return operators, nil
}

// Update persists the operator's mutable fields by id.
// It returns ErrOperatorNotFound when no row was affected.
func (r *PostgresOperatorRepository) Update(ctx context.Context, operator *operatordomain.Operator) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE operators
		SET name = $2,
		    active = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		operator.ID,
		operator.Name,
		operator.Active,
		operator.UpdatedAt,
		operator.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrOperatorNotFound
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// Unique violations become ErrOperatorNameAlreadyUsed; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrOperatorNameAlreadyUsed
	}
	return fmt.Errorf("No se pudo guardar el operador: %w", err)
}
