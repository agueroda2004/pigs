package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	userdomain "server/internal/modules/user/domain"
	"server/internal/modules/user/ports"
)

var ErrUserNotFound = ports.ErrUserNotFound

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a repository backed by a pgx pool.
// It stores the pool used for all user queries.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create inserts a new user row into the database.
// It maps unique constraint violations to ErrUsernameAlreadyUsed.
func (r *PostgresUserRepository) Create(ctx context.Context, user *userdomain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (
			id, name, username, password, role,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		user.ID,
		user.Name,
		user.Username,
		user.Password,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		nullableString(user.CreatedBy),
		nullableString(user.UpdatedBy),
	)
	if err != nil {
		return mapPostgresError(err)
	}
	return nil
}

// GetByID fetches a single user by its identifier.
// It returns ErrUserNotFound when no row matches.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error) {
	result := &userdomain.User{}
	var role string
	var createdBy, updatedBy *string

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, username, password, role,
		       created_at, updated_at, created_by, updated_by
		FROM users
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&result.Name,
		&result.Username,
		&result.Password,
		&role,
		&result.CreatedAt,
		&result.UpdatedAt,
		&createdBy,
		&updatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el usuario: %w", err)
	}

	result.Role = userdomain.Role(role)
	result.CreatedBy = stringValue(createdBy)
	result.UpdatedBy = stringValue(updatedBy)
	return result, nil
}

// FindByUsername fetches a single user by username.
// It returns ErrUserNotFound when no row matches.
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*userdomain.User, error) {
	result := &userdomain.User{}
	var role string
	var createdBy, updatedBy *string

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, username, password, role,
		       created_at, updated_at, created_by, updated_by
		FROM users
		WHERE username = $1
	`, username).Scan(
		&result.ID,
		&result.Name,
		&result.Username,
		&result.Password,
		&role,
		&result.CreatedAt,
		&result.UpdatedAt,
		&createdBy,
		&updatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el usuario: %w", err)
	}

	result.Role = userdomain.Role(role)
	result.CreatedBy = stringValue(createdBy)
	result.UpdatedBy = stringValue(updatedBy)
	return result, nil
}

// ExistsByUsername reports whether a username is already taken.
// It wraps any query failure with context.
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE username = $1
		)
	`, username).Scan(&exists); err != nil {
		return false, fmt.Errorf("No se pudo verificar el nombre de usuario: %w", err)
	}
	return exists, nil
}

// Update persists the user's mutable fields by id.
// It returns ErrUserNotFound when no row was affected.
func (r *PostgresUserRepository) Update(ctx context.Context, user *userdomain.User) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE users
		SET name = $2,
		    username = $3,
		    password = $4,
		    role = $5,
		    updated_at = $6,
		    updated_by = $7
		WHERE id = $1
	`,
		user.ID,
		user.Name,
		user.Username,
		user.Password,
		user.Role,
		user.UpdatedAt,
		nullableString(user.UpdatedBy),
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// Unique violations become ErrUsernameAlreadyUsed; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrUsernameAlreadyUsed
	}
	return fmt.Errorf("No se pudo guardar el usuario: %w", err)
}

// nullableString converts an empty string into a SQL NULL value.
// Non-empty values are returned unchanged.
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// stringValue dereferences a nullable string pointer.
// It returns an empty string when the pointer is nil.
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
