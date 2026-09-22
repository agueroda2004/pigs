package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "server/internal/modules/auth/domain"
	"server/internal/modules/auth/ports"
)

type PostgresRefreshTokenRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRefreshTokenRepository builds a refresh token repository backed by a pgx pool.
// It returns a repository ready to persist and query refresh tokens.
func NewPostgresRefreshTokenRepository(pool *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{pool: pool}
}

// Create inserts a new refresh token row into the database.
// It wraps any storage failure with a descriptive error.
func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *authdomain.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (
			id, user_id, family_id, token_hash, is_used, is_revoked, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		token.ID,
		token.UserID,
		token.FamilyID,
		token.TokenHash,
		token.IsUsed,
		token.IsRevoked,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el token de refresco: %w", err)
	}
	return nil
}

// GetByHash loads the refresh token matching the given hash.
// It returns ports.ErrRefreshTokenNotFound when no row matches.
func (r *PostgresRefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*authdomain.RefreshToken, error) {
	token := &authdomain.RefreshToken{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, family_id, token_hash, is_used, is_revoked, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.FamilyID,
		&token.TokenHash,
		&token.IsUsed,
		&token.IsRevoked,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el token de refresco: %w", err)
	}
	return token, nil
}

// MarkAsUsed marks the token as used only when it is still unused.
// It returns ports.ErrRefreshTokenAlreadyUsed when no row is updated.
func (r *PostgresRefreshTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET is_used = TRUE
		WHERE id = $1 AND is_used = FALSE
	`, id)
	if err != nil {
		return fmt.Errorf("No se pudo marcar el token como usado: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrRefreshTokenAlreadyUsed
	}
	return nil
}

// RevokeFamily revokes every non-revoked token belonging to the given family.
// It wraps any storage failure with a descriptive error.
func (r *PostgresRefreshTokenRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = TRUE
		WHERE family_id = $1 AND is_revoked = FALSE
	`, familyID)
	if err != nil {
		return fmt.Errorf("No se pudo revocar la familia de tokens: %w", err)
	}
	return nil
}

// RevokeByHash revokes the non-revoked token matching the given hash.
// It wraps any storage failure with a descriptive error.
func (r *PostgresRefreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = TRUE
		WHERE token_hash = $1 AND is_revoked = FALSE
	`, hash)
	if err != nil {
		return fmt.Errorf("No se pudo revocar el token de refresco: %w", err)
	}
	return nil
}
