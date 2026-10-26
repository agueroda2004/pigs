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
	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
	"server/internal/modules/pigletfostering/ports"
	sowdomain "server/internal/modules/sow/domain"
)

type PostgresPigletFosteringRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPigletFosteringRepository creates a repository backed by a pgx pool.
// It stores the pool used for all fostering, sow and farrowing queries.
func NewPostgresPigletFosteringRepository(pool *pgxpool.Pool) *PostgresPigletFosteringRepository {
	return &PostgresPigletFosteringRepository{pool: pool}
}

// GetSow fetches the sow referenced by a fostering by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresPigletFosteringRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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
func (r *PostgresPigletFosteringRepository) GetLastFarrowing(ctx context.Context, sowID uuid.UUID) (*farrowingdomain.Farrowing, error) {
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

// GetFarrowingByID fetches a farrowing by its identifier.
// It returns ports.ErrFarrowingNotFound when no row matches.
func (r *PostgresPigletFosteringRepository) GetFarrowingByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error) {
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

// GetLastPartialWeagingDate fetches the most recent partial weaging date of a farrowing.
// It orders by weaging date and returns the zero time when the farrowing has none.
func (r *PostgresPigletFosteringRepository) GetLastPartialWeagingDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error) {
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

// GetLastWeagingDate fetches the most recent full weaging date of a farrowing.
// It returns the zero time when the farrowing has not been weaned yet.
func (r *PostgresPigletFosteringRepository) GetLastWeagingDate(ctx context.Context, farrowingID uuid.UUID) (time.Time, error) {
	var weagingDate time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT weaging_date
		FROM weagings
		WHERE farrowing_id = $1
		ORDER BY weaging_date DESC, created_at DESC
		LIMIT 1
	`, farrowingID).Scan(&weagingDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("No se pudo consultar el último destete: %w", err)
	}
	return weagingDate, nil
}

// GetByID fetches a fostering by its identifier, joining both farrowings to
// expose the donor and receiver sows. It returns ports.ErrPigletFosteringNotFound
// when no row matches.
func (r *PostgresPigletFosteringRepository) GetByID(ctx context.Context, id uuid.UUID) (*pigletfosteringdomain.PigletFostering, error) {
	result := &pigletfosteringdomain.PigletFostering{}

	err := r.pool.QueryRow(ctx, `
		SELECT pf.id, pf.donor_farrowing_id, pf.receiver_farrowing_id,
		       df.sow_id, rf.sow_id, pf.movement_date, pf.quantity, pf.note,
		       pf.created_at, pf.updated_at, pf.created_by, pf.updated_by
		FROM piglet_fosterings pf
		JOIN farrowings df ON df.id = pf.donor_farrowing_id
		JOIN farrowings rf ON rf.id = pf.receiver_farrowing_id
		WHERE pf.id = $1
	`, id).Scan(
		&result.ID,
		&result.DonorFarrowingID,
		&result.ReceiverFarrowingID,
		&result.DonorSowID,
		&result.ReceiverSowID,
		&result.MovementDate,
		&result.Quantity,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrPigletFosteringNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el traslado de lechones: %w", err)
	}
	return result, nil
}

// Create inserts the fostering and moves the piglets between both farrowings.
// The donor decrement is guarded so it fails when the balance is insufficient,
// the receiver increment and the insert run in the same transaction.
func (r *PostgresPigletFosteringRepository) Create(
	ctx context.Context,
	fostering *pigletfosteringdomain.PigletFostering,
	donor *farrowingdomain.Farrowing,
	receiver *farrowingdomain.Farrowing,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el traslado de lechones: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets - $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1 AND current_piglets >= $2
	`,
		donor.ID,
		fostering.Quantity,
		donor.UpdatedAt,
		donor.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el traslado de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return farrowingdomain.ErrInsufficientPiglets
	}

	if _, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets + $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`,
		receiver.ID,
		fostering.Quantity,
		receiver.UpdatedAt,
		receiver.UpdatedBy,
	); err != nil {
		return fmt.Errorf("No se pudo guardar el traslado de lechones: %w", err)
	}

	if _, err := transaction.Exec(ctx, `
		INSERT INTO piglet_fosterings (
			id, donor_farrowing_id, receiver_farrowing_id, movement_date, quantity, note,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		fostering.ID,
		fostering.DonorFarrowingID,
		fostering.ReceiverFarrowingID,
		fostering.MovementDate,
		fostering.Quantity,
		fostering.Note,
		fostering.CreatedAt,
		fostering.UpdatedAt,
		fostering.CreatedBy,
		fostering.UpdatedBy,
	); err != nil {
		return fmt.Errorf("No se pudo guardar el traslado de lechones: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el traslado de lechones: %w", err)
	}
	return nil
}

// Update persists the editable fields of a fostering and rebalances the piglets.
// It locks the stored row, computes the quantity delta, applies it to both
// farrowings with guarded updates and rewrites date, quantity and note.
func (r *PostgresPigletFosteringRepository) Update(
	ctx context.Context,
	fostering *pigletfosteringdomain.PigletFostering,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var storedQuantity int
	err = transaction.QueryRow(ctx, `
		SELECT quantity
		FROM piglet_fosterings
		WHERE id = $1
		FOR UPDATE
	`, fostering.ID).Scan(&storedQuantity)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrPigletFosteringNotFound
	}
	if err != nil {
		return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
	}

	if err := updateFosteringBalances(
		ctx,
		transaction,
		fostering.DonorFarrowingID,
		fostering.ReceiverFarrowingID,
		fostering.Quantity-storedQuantity,
		fostering.UpdatedAt,
		fostering.UpdatedBy,
	); err != nil {
		return err
	}

	if _, err := transaction.Exec(ctx, `
		UPDATE piglet_fosterings
		SET movement_date = $2,
		    quantity = $3,
		    note = $4,
		    updated_at = $5,
		    updated_by = $6
		WHERE id = $1
	`,
		fostering.ID,
		fostering.MovementDate,
		fostering.Quantity,
		fostering.Note,
		fostering.UpdatedAt,
		fostering.UpdatedBy,
	); err != nil {
		return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
	}
	return nil
}

// Delete removes a fostering and undoes its balance movement in one transaction.
// It locks the stored row, takes the moved quantity back from the receiver with a
// guarded update, returns it to the donor and deletes the record.
func (r *PostgresPigletFosteringRepository) Delete(
	ctx context.Context,
	fostering *pigletfosteringdomain.PigletFostering,
	donor *farrowingdomain.Farrowing,
	receiver *farrowingdomain.Farrowing,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar el traslado de lechones: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var storedQuantity int
	err = transaction.QueryRow(ctx, `
		SELECT quantity
		FROM piglet_fosterings
		WHERE id = $1
		FOR UPDATE
	`, fostering.ID).Scan(&storedQuantity)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrPigletFosteringNotFound
	}
	if err != nil {
		return fmt.Errorf("No se pudo eliminar el traslado de lechones: %w", err)
	}

	if err := reduceFarrowingPiglets(
		ctx,
		transaction,
		receiver.ID,
		storedQuantity,
		receiver.UpdatedAt,
		receiver.UpdatedBy,
	); err != nil {
		return err
	}

	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets + $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`, donor.ID, storedQuantity, donor.UpdatedAt, donor.UpdatedBy)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar el traslado de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrFarrowingNotFound
	}

	commandTag, err = transaction.Exec(ctx, `
		DELETE FROM piglet_fosterings
		WHERE id = $1
	`, fostering.ID)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar el traslado de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrPigletFosteringNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo eliminar el traslado de lechones: %w", err)
	}
	return nil
}

// updateFosteringBalances applies the quantity delta to both farrowings.
// A positive delta subtracts from the donor and adds to the receiver, while a
// negative delta returns piglets to the donor and takes them back from the receiver.
func updateFosteringBalances(
	ctx context.Context,
	transaction pgx.Tx,
	donorFarrowingID uuid.UUID,
	receiverFarrowingID uuid.UUID,
	delta int,
	updatedAt time.Time,
	updatedBy uuid.UUID,
) error {
	if delta == 0 {
		return nil
	}

	if delta > 0 {
		if err := reduceFarrowingPiglets(ctx, transaction, donorFarrowingID, delta, updatedAt, updatedBy); err != nil {
			return err
		}
		_, err := transaction.Exec(ctx, `
			UPDATE farrowings
			SET current_piglets = current_piglets + $2,
			    updated_at = $3,
			    updated_by = $4
			WHERE id = $1
		`, receiverFarrowingID, delta, updatedAt, updatedBy)
		if err != nil {
			return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
		}
		return nil
	}

	amount := -delta
	if err := reduceFarrowingPiglets(ctx, transaction, receiverFarrowingID, amount, updatedAt, updatedBy); err != nil {
		return err
	}
	if _, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets + $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`, donorFarrowingID, amount, updatedAt, updatedBy); err != nil {
		return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
	}
	return nil
}

// reduceFarrowingPiglets subtracts a positive quantity from a farrowing balance.
// The update is guarded so it returns farrowingdomain.ErrInsufficientPiglets when
// the balance is not enough.
func reduceFarrowingPiglets(
	ctx context.Context,
	transaction pgx.Tx,
	farrowingID uuid.UUID,
	quantity int,
	updatedAt time.Time,
	updatedBy uuid.UUID,
) error {
	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET current_piglets = current_piglets - $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1 AND current_piglets >= $2
	`, farrowingID, quantity, updatedAt, updatedBy)
	if err != nil {
		return fmt.Errorf("No se pudo actualizar el traslado de lechones: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return farrowingdomain.ErrInsufficientPiglets
	}
	return nil
}

// List fetches the fosterings matching the filter, newest first.
// It joins both farrowings to expose the donor and receiver sows.
func (r *PostgresPigletFosteringRepository) List(ctx context.Context, filter ports.PigletFosteringFilter) ([]*pigletfosteringdomain.PigletFostering, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pf.id, pf.donor_farrowing_id, pf.receiver_farrowing_id,
		       df.sow_id, rf.sow_id, pf.movement_date, pf.quantity, pf.note,
		       pf.created_at, pf.updated_at, pf.created_by, pf.updated_by
		FROM piglet_fosterings pf
		JOIN farrowings df ON df.id = pf.donor_farrowing_id
		JOIN farrowings rf ON rf.id = pf.receiver_farrowing_id
		WHERE ($1::uuid IS NULL OR df.sow_id = $1)
		  AND ($2::uuid IS NULL OR rf.sow_id = $2)
		  AND ($3::date IS NULL OR pf.movement_date >= $3)
		  AND ($4::date IS NULL OR pf.movement_date <= $4)
		ORDER BY pf.movement_date DESC, pf.created_at DESC
	`, filter.DonorSowID, filter.ReceiverSowID, filter.FromDate, filter.ToDate)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los traslados de lechones: %w", err)
	}
	defer rows.Close()

	fosterings := make([]*pigletfosteringdomain.PigletFostering, 0)
	for rows.Next() {
		result := &pigletfosteringdomain.PigletFostering{}
		if err := rows.Scan(
			&result.ID,
			&result.DonorFarrowingID,
			&result.ReceiverFarrowingID,
			&result.DonorSowID,
			&result.ReceiverSowID,
			&result.MovementDate,
			&result.Quantity,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los traslados de lechones: %w", err)
		}
		fosterings = append(fosterings, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los traslados de lechones: %w", err)
	}
	return fosterings, nil
}
