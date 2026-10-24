package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	farrowingdomain "server/internal/modules/farrowing/domain"
	operatordomain "server/internal/modules/operator/domain"
	pigletdeathdomain "server/internal/modules/pigletdeath/domain"
	"server/internal/modules/pigletdeath/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type PostgresPigletDeathRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPigletDeathRepository creates a repository backed by a pgx pool.
// It stores the pool used for all piglet death, sow, farrowing and operator queries.
func NewPostgresPigletDeathRepository(pool *pgxpool.Pool) *PostgresPigletDeathRepository {
	return &PostgresPigletDeathRepository{pool: pool}
}

// GetSow fetches the sow referenced by a piglet death by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresPigletDeathRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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

// GetLastFarrowing fetches the most recent farrowing of a sow.
// It returns ports.ErrFarrowingNotFound when the sow has no farrowing.
func (r *PostgresPigletDeathRepository) GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
	result := &farrowingdomain.Farrowing{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, service_id, sow_id, farrow_date, start_time, end_time, location,
		       live_born, stillborn, mummified, current_piglets, litter_weight, stillborn_weight,
		       is_manipulated, is_nurse, nurse_start_date, note, created_at, updated_at, created_by, updated_by
		FROM farrowings
		WHERE sow_id = $1
		ORDER BY farrow_date DESC, created_at DESC
		LIMIT 1
	`, sowID).Scan(
		&result.ID,
		&result.ServiceID,
		&result.SowID,
		&result.FarrowDate,
		&result.StartTime,
		&result.EndTime,
		&result.Location,
		&result.LiveBorn,
		&result.Stillborn,
		&result.Mummified,
		&result.CurrentPiglets,
		&result.LitterWeight,
		&result.StillbornWeight,
		&result.IsManipulated,
		&result.IsNurse,
		&result.NurseStartDate,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrFarrowingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el parto: %w", err)
	}
	return result, nil
}

// GetLastPartialWeagingDate fetches the most recent partial weaging date of a farrowing.
// It orders by weaging date and returns the zero time when the farrowing has none.
func (r *PostgresPigletDeathRepository) GetLastPartialWeagingDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error) {
	var weagingDate time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT weaging_date
		FROM partial_weagings
		WHERE farrowing_id = $1
		ORDER BY weaging_date DESC, created_at DESC
		LIMIT 1
	`, farrowingID).Scan(&weagingDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("No se pudo consultar el último destete parcial: %w", err)
	}
	return weagingDate, nil
}

// GetOperator fetches an operator by its identifier.
// It returns ports.ErrOperatorNotFound when no row matches.
func (r *PostgresPigletDeathRepository) GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error) {
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

// Create inserts the piglet death and reduces the farrowing current piglets.
// The decrement is guarded so it fails when the balance is insufficient, and
// both writes run in a single transaction.
func (r *PostgresPigletDeathRepository) Create(
	ctx context.Context,
	death *pigletdeathdomain.PigletDeath,
	farrowing *farrowingdomain.Farrowing,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar la muerte de lechones: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets - $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1 AND current_piglets >= $2
	`,
		farrowing.ID,
		death.Quantity,
		farrowing.UpdatedAt,
		farrowing.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("No se pudo guardar la muerte de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return farrowingdomain.ErrInsufficientPiglets
	}

	if _, err := transaction.Exec(ctx, `
		INSERT INTO piglet_deaths (
			id, farrowing_id, operator_id, death_date, quantity, weight, cause, turn, note,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		death.ID,
		death.FarrowingID,
		death.OperatorID,
		death.DeathDate,
		death.Quantity,
		death.Weight,
		death.Cause,
		death.Turn,
		death.Note,
		death.CreatedAt,
		death.UpdatedAt,
		death.CreatedBy,
		death.UpdatedBy,
	); err != nil {
		return mapPostgresError(err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar la muerte de lechones: %w", err)
	}
	return nil
}

// GetFarrowingByID fetches a farrowing by its identifier.
// It returns ports.ErrFarrowingNotFound when no row matches.
func (r *PostgresPigletDeathRepository) GetFarrowingByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error) {
	result := &farrowingdomain.Farrowing{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, service_id, sow_id, farrow_date, start_time, end_time, location,
		       live_born, stillborn, mummified, current_piglets, litter_weight, stillborn_weight,
		       is_manipulated, is_nurse, nurse_start_date, note, created_at, updated_at, created_by, updated_by
		FROM farrowings
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&result.ServiceID,
		&result.SowID,
		&result.FarrowDate,
		&result.StartTime,
		&result.EndTime,
		&result.Location,
		&result.LiveBorn,
		&result.Stillborn,
		&result.Mummified,
		&result.CurrentPiglets,
		&result.LitterWeight,
		&result.StillbornWeight,
		&result.IsManipulated,
		&result.IsNurse,
		&result.NurseStartDate,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrFarrowingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el parto: %w", err)
	}
	return result, nil
}

// GetByID fetches a piglet death by its identifier.
// It joins the farrowing, sow and operator to fill the sow and operator read fields.
func (r *PostgresPigletDeathRepository) GetByID(ctx context.Context, id uuid.UUID) (*pigletdeathdomain.PigletDeath, error) {
	result := &pigletdeathdomain.PigletDeath{}

	err := r.pool.QueryRow(ctx, `
		SELECT pd.id, pd.farrowing_id, f.sow_id, s.code, pd.operator_id, o.name, pd.death_date,
		       pd.quantity, pd.weight, pd.cause, pd.turn, pd.note,
		       pd.created_at, pd.updated_at, pd.created_by, pd.updated_by
		FROM piglet_deaths pd
		JOIN farrowings f ON f.id = pd.farrowing_id
		JOIN sows s ON s.id = f.sow_id
		JOIN operators o ON o.id = pd.operator_id
		WHERE pd.id = $1
	`, id).Scan(
		&result.ID,
		&result.FarrowingID,
		&result.SowID,
		&result.SowCode,
		&result.OperatorID,
		&result.OperatorName,
		&result.DeathDate,
		&result.Quantity,
		&result.Weight,
		&result.Cause,
		&result.Turn,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrPigletDeathNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar la muerte de lechones: %w", err)
	}
	return result, nil
}

// HasWeagingAfter reports whether a full weaging exists strictly after the date.
// It compares at day granularity and returns false when there is no later weaging.
func (r *PostgresPigletDeathRepository) HasWeagingAfter(ctx context.Context, farrowingID uuid.UUID, date time.Time) (bool, error) {
	var exists bool

	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM weagings
			WHERE farrowing_id = $1 AND weaging_date > $2::date
		)
	`, farrowingID, date).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("No se pudo consultar el destete: %w", err)
	}
	return exists, nil
}

// HasPartialWeagingAfter reports whether a partial weaging exists strictly after the date.
// It compares at day granularity and returns false when there is no later partial weaging.
func (r *PostgresPigletDeathRepository) HasPartialWeagingAfter(ctx context.Context, farrowingID uuid.UUID, date time.Time) (bool, error) {
	var exists bool

	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM partial_weagings
			WHERE farrowing_id = $1 AND weaging_date > $2::date
		)
	`, farrowingID, date).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("No se pudo consultar el destete parcial: %w", err)
	}
	return exists, nil
}

// Update persists the editable fields of a piglet death.
// It returns ports.ErrPigletDeathNotFound when no row matches the identifier.
func (r *PostgresPigletDeathRepository) Update(ctx context.Context, death *pigletdeathdomain.PigletDeath) error {
	commandTag, err := r.pool.Exec(ctx, `
		UPDATE piglet_deaths
		SET operator_id = $2,
		    weight = $3,
		    cause = $4,
		    turn = $5,
		    note = $6,
		    updated_at = $7,
		    updated_by = $8
		WHERE id = $1
	`,
		death.ID,
		death.OperatorID,
		death.Weight,
		death.Cause,
		death.Turn,
		death.Note,
		death.UpdatedAt,
		death.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("No se pudo actualizar la muerte de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrPigletDeathNotFound
	}
	return nil
}

// Delete removes the piglet death and gives its quantity back to the farrowing.
// Both writes run in a single transaction so the rows always change together.
func (r *PostgresPigletDeathRepository) Delete(
	ctx context.Context,
	death *pigletdeathdomain.PigletDeath,
	farrowing *farrowingdomain.Farrowing,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar la muerte de lechones: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		DELETE FROM piglet_deaths
		WHERE id = $1
	`, death.ID)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar la muerte de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrPigletDeathNotFound
	}

	commandTag, err = transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets + $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`,
		farrowing.ID,
		death.Quantity,
		farrowing.UpdatedAt,
		farrowing.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar la muerte de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrFarrowingNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo eliminar la muerte de lechones: %w", err)
	}
	return nil
}

// List fetches one page of piglet deaths matching the filter, newest first, with
// the total count of matches. It joins the farrowing and sow to expose the sow code
// and returns an empty slice when no row matches.
func (r *PostgresPigletDeathRepository) List(
	ctx context.Context,
	filter ports.PigletDeathFilter,
	limit int,
	offset int,
) ([]*pigletdeathdomain.PigletDeath, int, error) {
	var total int

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM piglet_deaths pd
		JOIN farrowings f ON f.id = pd.farrowing_id
		JOIN sows s ON s.id = f.sow_id
		WHERE ($1::text IS NULL OR LOWER(s.code) = LOWER($1))
		  AND ($2::date IS NULL OR pd.death_date >= $2)
		  AND ($3::date IS NULL OR pd.death_date <= $3)
	`, filter.SowCode, filter.FromDate, filter.ToDate).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar las muertes de lechones: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT pd.id, pd.farrowing_id, f.sow_id, s.code, pd.operator_id, o.name, pd.death_date,
		       pd.quantity, pd.weight, pd.cause, pd.turn, pd.note,
		       pd.created_at, pd.updated_at, pd.created_by, pd.updated_by
		FROM piglet_deaths pd
		JOIN farrowings f ON f.id = pd.farrowing_id
		JOIN sows s ON s.id = f.sow_id
		JOIN operators o ON o.id = pd.operator_id
		WHERE ($1::text IS NULL OR LOWER(s.code) = LOWER($1))
		  AND ($2::date IS NULL OR pd.death_date >= $2)
		  AND ($3::date IS NULL OR pd.death_date <= $3)
		ORDER BY pd.death_date DESC, pd.created_at DESC
		LIMIT $4 OFFSET $5
	`, filter.SowCode, filter.FromDate, filter.ToDate, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar las muertes de lechones: %w", err)
	}
	defer rows.Close()

	deaths := make([]*pigletdeathdomain.PigletDeath, 0)
	for rows.Next() {
		result := &pigletdeathdomain.PigletDeath{}
		if err := rows.Scan(
			&result.ID,
			&result.FarrowingID,
			&result.SowID,
			&result.SowCode,
			&result.OperatorID,
			&result.OperatorName,
			&result.DeathDate,
			&result.Quantity,
			&result.Weight,
			&result.Cause,
			&result.Turn,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, 0, fmt.Errorf("No se pudieron consultar las muertes de lechones: %w", err)
		}
		deaths = append(deaths, result)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar las muertes de lechones: %w", err)
	}
	return deaths, total, nil
}

// mapPostgresError wraps a PostgreSQL write failure with context.
// It keeps the message used by the module for insertion errors.
func mapPostgresError(err error) error {
	return fmt.Errorf("No se pudo guardar la muerte de lechones: %w", err)
}
