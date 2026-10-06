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

	farrowingdomain "server/internal/modules/farrowing/domain"
	sowdomain "server/internal/modules/sow/domain"
	weagingdomain "server/internal/modules/weaging/domain"
	"server/internal/modules/weaging/ports"
)

type PostgresWeagingRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresWeagingRepository creates a repository backed by a pgx pool.
// It stores the pool used for all weaging, sow and farrowing queries.
func NewPostgresWeagingRepository(pool *pgxpool.Pool) *PostgresWeagingRepository {
	return &PostgresWeagingRepository{pool: pool}
}

// GetSow fetches the sow referenced by a weaging by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresWeagingRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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
func (r *PostgresWeagingRepository) GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
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

// GetLatestEventDate returns the most recent death, fostering or partial weaging date
// of a farrowing. It returns the zero time when the farrowing has no related events.
func (r *PostgresWeagingRepository) GetLatestEventDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error) {
	var lastEventDate time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(
			GREATEST(
				(SELECT MAX(death_date) FROM piglet_deaths WHERE farrowing_id = $1),
				(SELECT MAX(movement_date) FROM piglet_fosterings
				 WHERE donor_farrowing_id = $1 OR receiver_farrowing_id = $1),
				(SELECT MAX(weaging_date) FROM partial_weagings WHERE farrowing_id = $1)
			),
			'0001-01-01'::date
		)
	`, farrowingID).Scan(&lastEventDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("No se pudieron consultar los eventos del parto: %w", err)
	}
	return lastEventDate, nil
}

// Create inserts the weaging, zeroes the farrowing current piglets and weans the sow.
// The update is guarded so it fails when the balance changed, and every write runs in
// one transaction. A duplicate farrowing weaging maps to ports.ErrWeagingAlreadyExists.
func (r *PostgresWeagingRepository) Create(
	ctx context.Context,
	weaging *weagingdomain.Weaging,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el destete: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = 0,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1 AND current_piglets = $2
	`,
		farrowing.ID,
		weaging.Quantity,
		farrowing.UpdatedAt,
		farrowing.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el destete: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return farrowingdomain.ErrWeagingQuantityMismatch
	}

	commandTag, err = transaction.Exec(ctx, `
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
		return fmt.Errorf("No se pudo guardar el destete: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrSowNotFound
	}

	if _, err := transaction.Exec(ctx, `
		INSERT INTO weagings (
			id, farrowing_id, weaging_date, quantity, total_weight, destination, note,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		weaging.ID,
		weaging.FarrowingID,
		weaging.WeagingDate,
		weaging.Quantity,
		weaging.TotalWeight,
		weaging.Destination,
		weaging.Note,
		weaging.CreatedAt,
		weaging.UpdatedAt,
		weaging.CreatedBy,
		weaging.UpdatedBy,
	); err != nil {
		return mapPostgresError(err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el destete: %w", err)
	}
	return nil
}

// List fetches the weagings matching the filter, newest first.
// It joins the farrowing to expose the sow and returns an empty slice when no row matches.
func (r *PostgresWeagingRepository) List(ctx context.Context, filter ports.WeagingFilter) ([]*weagingdomain.Weaging, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT w.id, w.farrowing_id, f.sow_id, w.weaging_date, w.quantity, w.total_weight,
		       w.destination, w.note, w.created_at, w.updated_at, w.created_by, w.updated_by
		FROM weagings w
		JOIN farrowings f ON f.id = w.farrowing_id
		WHERE ($1::uuid IS NULL OR f.sow_id = $1)
		  AND ($2::date IS NULL OR w.weaging_date >= $2)
		  AND ($3::date IS NULL OR w.weaging_date <= $3)
		ORDER BY w.weaging_date DESC, w.created_at DESC
	`, filter.SowID, filter.FromDate, filter.ToDate)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los destetes: %w", err)
	}
	defer rows.Close()

	weagings := make([]*weagingdomain.Weaging, 0)
	for rows.Next() {
		result := &weagingdomain.Weaging{}
		if err := rows.Scan(
			&result.ID,
			&result.FarrowingID,
			&result.SowID,
			&result.WeagingDate,
			&result.Quantity,
			&result.TotalWeight,
			&result.Destination,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los destetes: %w", err)
		}
		weagings = append(weagings, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los destetes: %w", err)
	}
	return weagings, nil
}

// mapPostgresError translates PostgreSQL errors into port errors.
// A unique violation means the farrowing already has a weaging; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrWeagingAlreadyExists
	}
	return fmt.Errorf("No se pudo guardar el destete: %w", err)
}
