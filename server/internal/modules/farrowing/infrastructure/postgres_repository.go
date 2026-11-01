package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	farrowingdomain "server/internal/modules/farrowing/domain"
	"server/internal/modules/farrowing/ports"
	medicationdomain "server/internal/modules/medication/domain"
	operatordomain "server/internal/modules/operator/domain"
	servicedomain "server/internal/modules/service/domain"
	sowdomain "server/internal/modules/sow/domain"
)

type PostgresFarrowingRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresFarrowingRepository creates a repository backed by a pgx pool.
// It stores the pool used for all farrowing, sow, service and lookup queries.
func NewPostgresFarrowingRepository(pool *pgxpool.Pool) *PostgresFarrowingRepository {
	return &PostgresFarrowingRepository{pool: pool}
}

// GetSow fetches the sow referenced by a farrowing by its identifier.
// It returns ports.ErrSowNotFound when no row matches.
func (r *PostgresFarrowingRepository) GetSow(ctx context.Context, id uuid.UUID) (*sowdomain.Sow, error) {
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

// GetLastService fetches the most recent service of a sow along with its mounts.
// It returns ports.ErrServiceNotFound when the sow has no service.
func (r *PostgresFarrowingRepository) GetLastService(ctx context.Context, sowID uuid.UUID) (*servicedomain.Service, error) {
	result := &servicedomain.Service{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE sow_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, sowID).Scan(
		&result.ID,
		&result.SowID,
		&result.ExpectedFarrowingDate,
		&result.Note,
		&result.State,
		&result.Location,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrServiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el servicio: %w", err)
	}

	result.Mounts, err = r.listMounts(ctx, result.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetByID fetches a farrowing by its identifier with its sow code, operators
// and medications loaded. It returns ports.ErrFarrowingNotFound when missing.
func (r *PostgresFarrowingRepository) GetByID(ctx context.Context, id uuid.UUID) (*farrowingdomain.Farrowing, error) {
	result := &farrowingdomain.Farrowing{}

	err := r.pool.QueryRow(ctx, `
		SELECT f.id, f.service_id, f.sow_id, s.code, f.farrow_date, f.start_time, f.end_time,
		       f.location, f.live_born, f.stillborn, f.mummified, f.current_piglets,
		       f.litter_weight, f.stillborn_weight, f.is_manipulated, f.is_nurse,
		       f.nurse_start_date, f.note, f.created_at, f.updated_at, f.created_by, f.updated_by
		FROM farrowings f
		JOIN sows s ON s.id = f.sow_id
		WHERE f.id = $1
	`, id).Scan(
		&result.ID,
		&result.ServiceID,
		&result.SowID,
		&result.SowCode,
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

	if result.Operators, err = r.listOperators(ctx, result.ID); err != nil {
		return nil, err
	}
	if result.Medications, err = r.listMedications(ctx, result.ID); err != nil {
		return nil, err
	}
	return result, nil
}

// GetService fetches a service by its identifier with its mounts loaded.
// It returns ports.ErrServiceNotFound when no row matches.
func (r *PostgresFarrowingRepository) GetService(ctx context.Context, id uuid.UUID) (*servicedomain.Service, error) {
	result := &servicedomain.Service{}

	err := r.pool.QueryRow(ctx, `
		SELECT id, sow_id, expected_farrowing_date, note, state, location,
		       created_at, updated_at, created_by, updated_by
		FROM services
		WHERE id = $1
	`, id).Scan(
		&result.ID,
		&result.SowID,
		&result.ExpectedFarrowingDate,
		&result.Note,
		&result.State,
		&result.Location,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrServiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el servicio: %w", err)
	}

	result.Mounts, err = r.listMounts(ctx, result.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetEarliestRelatedDate returns the earliest date among the deaths, fosterings,
// partial weagings and weagings linked to the farrowing. It returns nil when the
// farrowing has none of those records.
func (r *PostgresFarrowingRepository) GetEarliestRelatedDate(ctx context.Context, farrowingID uuid.UUID) (*time.Time, error) {
	var earliest *time.Time
	if err := r.pool.QueryRow(ctx, `
		SELECT LEAST(
			(SELECT MIN(death_date) FROM piglet_deaths WHERE farrowing_id = $1),
			(SELECT MIN(movement_date) FROM piglet_fosterings
			 WHERE donor_farrowing_id = $1 OR receiver_farrowing_id = $1),
			(SELECT MIN(weaging_date) FROM partial_weagings WHERE farrowing_id = $1),
			(SELECT MIN(weaging_date) FROM weagings WHERE farrowing_id = $1)
		)
	`, farrowingID).Scan(&earliest); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los registros del parto: %w", err)
	}
	return earliest, nil
}

// listMounts fetches every mount of a service ordered by mount number.
// It returns an empty slice when the service has no mounts.
func (r *PostgresFarrowingRepository) listMounts(ctx context.Context, serviceID uuid.UUID) ([]*servicedomain.Mount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, service_id, boar_id, operator_id, mount_number, mount_date,
		       type, note, created_at, updated_at, created_by, updated_by
		FROM mounts
		WHERE service_id = $1
		ORDER BY mount_number ASC
	`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}
	defer rows.Close()

	mounts := make([]*servicedomain.Mount, 0)
	for rows.Next() {
		result := &servicedomain.Mount{}
		if err := rows.Scan(
			&result.ID,
			&result.ServiceID,
			&result.BoarID,
			&result.OperatorID,
			&result.MountNumber,
			&result.MountDate,
			&result.Type,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
		}
		mounts = append(mounts, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar las montas: %w", err)
	}
	return mounts, nil
}

// GetOperator fetches an operator by its identifier.
// It returns ports.ErrOperatorNotFound when no row matches.
func (r *PostgresFarrowingRepository) GetOperator(ctx context.Context, id uuid.UUID) (*operatordomain.Operator, error) {
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

// GetMedication fetches a medication by its identifier.
// It returns ports.ErrMedicationNotFound when no row matches.
func (r *PostgresFarrowingRepository) GetMedication(ctx context.Context, id uuid.UUID) (*medicationdomain.Medication, error) {
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
		return nil, ports.ErrMedicationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("No se pudo consultar el medicamento: %w", err)
	}
	return result, nil
}

// Create inserts the farrowing, its operator and medication links and moves the
// sow to lactating and the service to finished.
// All writes run in a single transaction so every row changes together.
func (r *PostgresFarrowingRepository) Create(
	ctx context.Context,
	farrowing *farrowingdomain.Farrowing,
	sow *sowdomain.Sow,
	service *servicedomain.Service,
) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el parto: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	if _, err := transaction.Exec(ctx, `
		INSERT INTO farrowings (
			id, service_id, sow_id, farrow_date, start_time, end_time, location,
			live_born, stillborn, mummified, current_piglets, litter_weight, stillborn_weight,
			is_manipulated, is_nurse, nurse_start_date, note, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21
		)
	`,
		farrowing.ID,
		farrowing.ServiceID,
		farrowing.SowID,
		farrowing.FarrowDate,
		farrowing.StartTime,
		farrowing.EndTime,
		farrowing.Location,
		farrowing.LiveBorn,
		farrowing.Stillborn,
		farrowing.Mummified,
		farrowing.CurrentPiglets,
		farrowing.LitterWeight,
		farrowing.StillbornWeight,
		farrowing.IsManipulated,
		farrowing.IsNurse,
		farrowing.NurseStartDate,
		farrowing.Note,
		farrowing.CreatedAt,
		farrowing.UpdatedAt,
		farrowing.CreatedBy,
		farrowing.UpdatedBy,
	); err != nil {
		return mapPostgresError(err)
	}

	for _, operator := range farrowing.Operators {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO farrowing_operators (id, farrowing_id, operator_id, created_at)
			VALUES ($1, $2, $3, $4)
		`, operator.ID, operator.FarrowingID, operator.OperatorID, operator.CreatedAt); err != nil {
			return mapPostgresError(err)
		}
	}

	for _, medication := range farrowing.Medications {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO farrowing_medications (
				id, farrowing_id, medication_id, dose, applied_by, created_at
			) VALUES ($1, $2, $3, $4, $5, $6)
		`,
			medication.ID,
			medication.FarrowingID,
			medication.MedicationID,
			medication.Dose,
			medication.AppliedBy,
			medication.CreatedAt,
		); err != nil {
			return mapPostgresError(err)
		}
	}

	commandTag, err := transaction.Exec(ctx, `
		UPDATE sows
		SET state = $2,
		    parity = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		sow.ID,
		sow.State,
		sow.Parity,
		sow.UpdatedAt,
		sow.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrSowNotFound
	}

	commandTag, err = transaction.Exec(ctx, `
		UPDATE services
		SET state = $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`,
		service.ID,
		service.State,
		service.UpdatedAt,
		service.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrServiceNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el parto: %w", err)
	}
	return nil
}

// Update persists the editable fields of a farrowing and optionally replaces its
// operator and medication links. All writes run in a single transaction.
func (r *PostgresFarrowingRepository) Update(ctx context.Context, farrowing *farrowingdomain.Farrowing, replaceOperators, replaceMedications bool) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo guardar el parto: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE farrowings
		SET farrow_date = $2,
		    start_time = $3,
		    end_time = $4,
		    location = $5,
		    live_born = $6,
		    stillborn = $7,
		    mummified = $8,
		    current_piglets = $9,
		    litter_weight = $10,
		    stillborn_weight = $11,
		    is_manipulated = $12,
		    note = $13,
		    updated_at = $14,
		    updated_by = $15
		WHERE id = $1
	`,
		farrowing.ID,
		farrowing.FarrowDate,
		farrowing.StartTime,
		farrowing.EndTime,
		farrowing.Location,
		farrowing.LiveBorn,
		farrowing.Stillborn,
		farrowing.Mummified,
		farrowing.CurrentPiglets,
		farrowing.LitterWeight,
		farrowing.StillbornWeight,
		farrowing.IsManipulated,
		farrowing.Note,
		farrowing.UpdatedAt,
		farrowing.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrFarrowingNotFound
	}

	if replaceOperators {
		if _, err := transaction.Exec(ctx, `
			DELETE FROM farrowing_operators
			WHERE farrowing_id = $1
		`, farrowing.ID); err != nil {
			return mapPostgresError(err)
		}
		for _, operator := range farrowing.Operators {
			if _, err := transaction.Exec(ctx, `
				INSERT INTO farrowing_operators (id, farrowing_id, operator_id, created_at)
				VALUES ($1, $2, $3, $4)
			`, operator.ID, operator.FarrowingID, operator.OperatorID, operator.CreatedAt); err != nil {
				return mapPostgresError(err)
			}
		}
	}

	if replaceMedications {
		if _, err := transaction.Exec(ctx, `
			DELETE FROM farrowing_medications
			WHERE farrowing_id = $1
		`, farrowing.ID); err != nil {
			return mapPostgresError(err)
		}
		for _, medication := range farrowing.Medications {
			if _, err := transaction.Exec(ctx, `
				INSERT INTO farrowing_medications (
					id, farrowing_id, medication_id, dose, applied_by, created_at
				) VALUES ($1, $2, $3, $4, $5, $6)
			`,
				medication.ID,
				medication.FarrowingID,
				medication.MedicationID,
				medication.Dose,
				medication.AppliedBy,
				medication.CreatedAt,
			); err != nil {
				return mapPostgresError(err)
			}
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo guardar el parto: %w", err)
	}
	return nil
}

// Delete removes a farrowing and, in the same transaction, persists the restored
// sow state (including its parity) and service state.
// A linked-records foreign key violation becomes ports.ErrFarrowingHasRelatedRecords.
func (r *PostgresFarrowingRepository) Delete(ctx context.Context, farrowing *farrowingdomain.Farrowing, sow *sowdomain.Sow, service *servicedomain.Service) error {
	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("No se pudo eliminar el parto: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	commandTag, err := transaction.Exec(ctx, `
		UPDATE sows
		SET state = $2,
		    parity = $3,
		    updated_at = $4,
		    updated_by = $5
		WHERE id = $1
	`,
		sow.ID,
		sow.State,
		sow.Parity,
		sow.UpdatedAt,
		sow.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrSowNotFound
	}

	commandTag, err = transaction.Exec(ctx, `
		UPDATE services
		SET state = $2,
		    updated_at = $3,
		    updated_by = $4
		WHERE id = $1
	`,
		service.ID,
		service.State,
		service.UpdatedAt,
		service.UpdatedBy,
	)
	if err != nil {
		return mapPostgresError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrServiceNotFound
	}

	commandTag, err = transaction.Exec(ctx, `
		DELETE FROM farrowings
		WHERE id = $1
	`, farrowing.ID)
	if err != nil {
		return mapDeleteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrFarrowingNotFound
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("No se pudo eliminar el parto: %w", err)
	}
	return nil
}

// mapDeleteError translates PostgreSQL errors into port errors.
// A linked-records foreign key violation becomes ports.ErrFarrowingHasRelatedRecords
// and any other error is wrapped with context.
func mapDeleteError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503":
			return ports.ErrFarrowingHasRelatedRecords
		}
	}
	return fmt.Errorf("No se pudo eliminar el parto: %w", err)
}

// List fetches one page of farrowings matching the filter, newest first, with
// their sow code, operators and medications loaded. It returns the items, the
// total number of matches and an error.
func (r *PostgresFarrowingRepository) List(ctx context.Context, filter ports.FarrowingFilter, limit, offset int) ([]*farrowingdomain.Farrowing, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM farrowings f
		JOIN sows s ON s.id = f.sow_id
		WHERE ($1::text IS NULL OR LOWER(s.code) = LOWER($1))
		  AND ($2::date IS NULL OR f.farrow_date >= $2)
		  AND ($3::date IS NULL OR f.farrow_date <= $3)
	`, filter.SowCode, filter.FromDate, filter.ToDate).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT f.id, f.service_id, f.sow_id, s.code, f.farrow_date, f.start_time, f.end_time,
		       f.location, f.live_born, f.stillborn, f.mummified, f.current_piglets,
		       f.litter_weight, f.stillborn_weight, f.is_manipulated, f.is_nurse,
		       f.nurse_start_date, f.note, f.created_at, f.updated_at, f.created_by, f.updated_by
		FROM farrowings f
		JOIN sows s ON s.id = f.sow_id
		WHERE ($1::text IS NULL OR LOWER(s.code) = LOWER($1))
		  AND ($2::date IS NULL OR f.farrow_date >= $2)
		  AND ($3::date IS NULL OR f.farrow_date <= $3)
		ORDER BY f.farrow_date DESC, f.created_at DESC
		LIMIT $4 OFFSET $5
	`, filter.SowCode, filter.FromDate, filter.ToDate, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}
	defer rows.Close()

	farrowings := make([]*farrowingdomain.Farrowing, 0)
	for rows.Next() {
		result := &farrowingdomain.Farrowing{}
		if err := rows.Scan(
			&result.ID,
			&result.ServiceID,
			&result.SowID,
			&result.SowCode,
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
		); err != nil {
			return nil, 0, fmt.Errorf("No se pudieron consultar los partos: %w", err)
		}
		farrowings = append(farrowings, result)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}

	for _, farrowing := range farrowings {
		if farrowing.Operators, err = r.listOperators(ctx, farrowing.ID); err != nil {
			return nil, 0, err
		}
		if farrowing.Medications, err = r.listMedications(ctx, farrowing.ID); err != nil {
			return nil, 0, err
		}
	}
	return farrowings, total, nil
}

// HistoryBySow loads every farrowing of a sow with its operator and medication names.
// It returns an empty slice when the sow has no farrowings and never duplicates events.
func (r *PostgresFarrowingRepository) HistoryBySow(ctx context.Context, sowCode string) ([]*farrowingdomain.FarrowingHistoryBySow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT f.id, f.farrow_date, f.start_time, f.end_time, f.location,
		       f.live_born, f.mummified, f.stillborn, f.litter_weight,
		       f.stillborn_weight, f.is_manipulated, f.note, f.is_nurse
		FROM farrowings f
		JOIN sows s ON s.id = f.sow_id
		WHERE LOWER(s.code) = LOWER($1)
		ORDER BY f.farrow_date DESC, f.created_at DESC
	`, strings.TrimSpace(sowCode))
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}
	defer rows.Close()

	items := make([]*farrowingdomain.FarrowingHistoryBySow, 0)
	ids := make([]string, 0)
	for rows.Next() {
		item := &farrowingdomain.FarrowingHistoryBySow{}
		var farrowDate time.Time
		if err := rows.Scan(
			&item.ID,
			&farrowDate,
			&item.StartTime,
			&item.EndTime,
			&item.Location,
			&item.LiveBorn,
			&item.Mummified,
			&item.Stillborn,
			&item.LitterWeight,
			&item.StillbornWeight,
			&item.IsManipulated,
			&item.Note,
			&item.IsNurse,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los partos: %w", err)
		}
		item.FarrowDate = farrowDate.UTC().Format("2006-01-02")
		item.Operators = make([]farrowingdomain.FarrowingHistoryOperator, 0)
		item.Medications = make([]farrowingdomain.FarrowingHistoryMedication, 0)
		items = append(items, item)
		ids = append(ids, item.ID.String())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}
	if len(items) == 0 {
		return items, nil
	}

	operators, err := r.historyOperators(ctx, ids)
	if err != nil {
		return nil, err
	}
	medications, err := r.historyMedications(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if values, ok := operators[item.ID]; ok {
			item.Operators = values
		}
		if values, ok := medications[item.ID]; ok {
			item.Medications = values
		}
	}
	return items, nil
}

// historyOperators loads the operator names of the given farrowings grouped by farrowing id.
// It returns an empty map when none of the farrowings has an operator.
func (r *PostgresFarrowingRepository) historyOperators(ctx context.Context, farrowingIDs []string) (map[uuid.UUID][]farrowingdomain.FarrowingHistoryOperator, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT fo.farrowing_id, o.name
		FROM farrowing_operators fo
		JOIN operators o ON o.id = fo.operator_id
		WHERE fo.farrowing_id = ANY($1::uuid[])
		ORDER BY fo.farrowing_id, fo.created_at ASC
	`, farrowingIDs)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los operadores del parto: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]farrowingdomain.FarrowingHistoryOperator, len(farrowingIDs))
	for rows.Next() {
		var farrowingID uuid.UUID
		var name string
		if err := rows.Scan(&farrowingID, &name); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los operadores del parto: %w", err)
		}
		result[farrowingID] = append(result[farrowingID], farrowingdomain.FarrowingHistoryOperator{OperatorName: name})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los operadores del parto: %w", err)
	}
	return result, nil
}

// historyMedications loads the medication names, doses and applying operators of the
// given farrowings grouped by farrowing id. It returns an empty map when none exist.
func (r *PostgresFarrowingRepository) historyMedications(ctx context.Context, farrowingIDs []string) (map[uuid.UUID][]farrowingdomain.FarrowingHistoryMedication, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT fm.farrowing_id, m.name, fm.dose, o.name
		FROM farrowing_medications fm
		JOIN medications m ON m.id = fm.medication_id
		JOIN operators o ON o.id = fm.applied_by
		WHERE fm.farrowing_id = ANY($1::uuid[])
		ORDER BY fm.farrowing_id, fm.created_at ASC
	`, farrowingIDs)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los medicamentos del parto: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]farrowingdomain.FarrowingHistoryMedication, len(farrowingIDs))
	for rows.Next() {
		var farrowingID uuid.UUID
		var medication farrowingdomain.FarrowingHistoryMedication
		if err := rows.Scan(&farrowingID, &medication.MedicationName, &medication.Dose, &medication.AppliedByName); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los medicamentos del parto: %w", err)
		}
		result[farrowingID] = append(result[farrowingID], medication)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los medicamentos del parto: %w", err)
	}
	return result, nil
}

// listOperators fetches the operator links of a farrowing.
// It returns an empty slice when the farrowing has no operators.
func (r *PostgresFarrowingRepository) listOperators(ctx context.Context, farrowingID uuid.UUID) ([]*farrowingdomain.FarrowingOperator, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, farrowing_id, operator_id, created_at
		FROM farrowing_operators
		WHERE farrowing_id = $1
		ORDER BY created_at ASC
	`, farrowingID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los operadores del parto: %w", err)
	}
	defer rows.Close()

	operators := make([]*farrowingdomain.FarrowingOperator, 0)
	for rows.Next() {
		result := &farrowingdomain.FarrowingOperator{}
		if err := rows.Scan(&result.ID, &result.FarrowingID, &result.OperatorID, &result.CreatedAt); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los operadores del parto: %w", err)
		}
		operators = append(operators, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los operadores del parto: %w", err)
	}
	return operators, nil
}

// listMedications fetches the medication links of a farrowing.
// It returns an empty slice when the farrowing has no medications.
func (r *PostgresFarrowingRepository) listMedications(ctx context.Context, farrowingID uuid.UUID) ([]*farrowingdomain.FarrowingMedication, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, farrowing_id, medication_id, dose, applied_by, created_at
		FROM farrowing_medications
		WHERE farrowing_id = $1
		ORDER BY created_at ASC
	`, farrowingID)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los medicamentos del parto: %w", err)
	}
	defer rows.Close()

	medications := make([]*farrowingdomain.FarrowingMedication, 0)
	for rows.Next() {
		result := &farrowingdomain.FarrowingMedication{}
		if err := rows.Scan(
			&result.ID,
			&result.FarrowingID,
			&result.MedicationID,
			&result.Dose,
			&result.AppliedBy,
			&result.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los medicamentos del parto: %w", err)
		}
		medications = append(medications, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los medicamentos del parto: %w", err)
	}
	return medications, nil
}

// mapPostgresError translates PostgreSQL errors into domain port errors.
// A unique violation means the service already has a farrowing; others are wrapped.
func mapPostgresError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ports.ErrFarrowingAlreadyExists
	}
	return fmt.Errorf("No se pudo guardar el parto: %w", err)
}
