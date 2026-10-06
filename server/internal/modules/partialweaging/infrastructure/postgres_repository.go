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
	partialweagingdomain "server/internal/modules/partialweaging/domain"
	"server/internal/modules/partialweaging/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type PostgresPartialWeagingRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPartialWeagingRepository creates a repository backed by a pgx pool.
// It stores the pool used for all partial weaging, sow and farrowing queries.
func NewPostgresPartialWeagingRepository(pool *pgxpool.Pool) *PostgresPartialWeagingRepository {
	return &PostgresPartialWeagingRepository{pool: pool}
}

// GetSow fetches the sow referenced by a partial weaging by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresPartialWeagingRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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
func (r *PostgresPartialWeagingRepository) GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
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

// GetLatestEventDate returns the most recent death or fostering date of a farrowing.
// It returns the zero time when the farrowing has no related events.
func (r *PostgresPartialWeagingRepository) GetLatestEventDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error) {
	var lastEventDate time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(
			GREATEST(
				(SELECT MAX(death_date) FROM piglet_deaths WHERE farrowing_id = $1),
				(SELECT MAX(movement_date) FROM piglet_fosterings
				 WHERE donor_farrowing_id = $1 OR receiver_farrowing_id = $1)
			),
			'0001-01-01'::date
		)
	`, farrowingID).Scan(&lastEventDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("No se pudieron consultar los eventos del parto: %w", err)
	}
	return lastEventDate, nil
}

// Create inserts the partial weaging, reduces the farrowing current piglets and
// updates the sow state and the farrowing nurse flags. The decrement is guarded so
// it fails when the balance is insufficient and every write runs in one transaction.
func (r *PostgresPartialWeagingRepository) Create(
	ctx context.Context,
	weaging *partialweagingdomain.PartialWeaging,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el destete parcial: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets - $2,
		    is_nurse = $5,
		    nurse_start_date = $6,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1 AND current_piglets >= $2
	`,
		farrowing.ID,
		weaging.Quantity,
		farrowing.UpdatedAt,
		farrowing.UpdatedBy,
		farrowing.IsNurse,
		farrowing.NurseStartDate,
	)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el destete parcial: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return farrowingdomain.ErrInsufficientPiglets
	}

	if weaging.Type != partialweagingdomain.TypeNodriza {
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
			return fmt.Errorf("No se pudo guardar el destete parcial: %w", err)
		}
		if commandTag.RowsAffected() == 0 {
			return ports.ErrSowNotFound
		}
	}

	if _, err := transaction.Exec(ctx, `
		INSERT INTO partial_weagings (
			id, farrowing_id, weaging_date, quantity, total_weight, type, note,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		weaging.ID,
		weaging.FarrowingID,
		weaging.WeagingDate,
		weaging.Quantity,
		weaging.TotalWeight,
		weaging.Type,
		weaging.Note,
		weaging.CreatedAt,
		weaging.UpdatedAt,
		weaging.CreatedBy,
		weaging.UpdatedBy,
	); err != nil {
		return fmt.Errorf("No se pudo guardar el destete parcial: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el destete parcial: %w", err)
	}
	return nil
}

// List fetches the partial weagings matching the filter, newest first.
// It joins the farrowing to expose the sow and returns an empty slice when no row matches.
func (r *PostgresPartialWeagingRepository) List(ctx context.Context, filter ports.PartialWeagingFilter) ([]*partialweagingdomain.PartialWeaging, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pw.id, pw.farrowing_id, f.sow_id, pw.weaging_date, pw.quantity, pw.total_weight,
		       pw.type, pw.note, pw.created_at, pw.updated_at, pw.created_by, pw.updated_by
		FROM partial_weagings pw
		JOIN farrowings f ON f.id = pw.farrowing_id
		WHERE ($1::uuid IS NULL OR f.sow_id = $1)
		  AND ($2::date IS NULL OR pw.weaging_date >= $2)
		  AND ($3::date IS NULL OR pw.weaging_date <= $3)
		ORDER BY pw.weaging_date DESC, pw.created_at DESC
	`, filter.SowID, filter.FromDate, filter.ToDate)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los destetes parciales: %w", err)
	}
	defer rows.Close()

	weagings := make([]*partialweagingdomain.PartialWeaging, 0)
	for rows.Next() {
		result := &partialweagingdomain.PartialWeaging{}
		if err := rows.Scan(
			&result.ID,
			&result.FarrowingID,
			&result.SowID,
			&result.WeagingDate,
			&result.Quantity,
			&result.TotalWeight,
			&result.Type,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los destetes parciales: %w", err)
		}
		weagings = append(weagings, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los destetes parciales: %w", err)
	}
	return weagings, nil
}
