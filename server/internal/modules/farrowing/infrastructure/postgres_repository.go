package infrastructure

import (
	"context"
	"errors"
	"fmt"

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
			live_born, stillborn, mummified, litter_weight, stillborn_weight,
			is_manipulated, note, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18
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
		farrowing.LitterWeight,
		farrowing.StillbornWeight,
		farrowing.IsManipulated,
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

// List fetches the farrowings matching the filter, newest first, with their
// operators and medications loaded.
// It returns an empty slice when no farrowing matches.
func (r *PostgresFarrowingRepository) List(ctx context.Context, filter ports.FarrowingFilter) ([]*farrowingdomain.Farrowing, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, service_id, sow_id, farrow_date, start_time, end_time, location,
		       live_born, stillborn, mummified, litter_weight, stillborn_weight,
		       is_manipulated, note, created_at, updated_at, created_by, updated_by
		FROM farrowings
		WHERE ($1::uuid IS NULL OR sow_id = $1)
		  AND ($2::uuid IS NULL OR service_id = $2)
		  AND ($3::date IS NULL OR farrow_date >= $3)
		  AND ($4::date IS NULL OR farrow_date <= $4)
		ORDER BY farrow_date DESC, created_at DESC
	`, filter.SowID, filter.ServiceID, filter.FromDate, filter.ToDate)
	if err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}
	defer rows.Close()

	farrowings := make([]*farrowingdomain.Farrowing, 0)
	for rows.Next() {
		result := &farrowingdomain.Farrowing{}
		if err := rows.Scan(
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
			&result.LitterWeight,
			&result.StillbornWeight,
			&result.IsManipulated,
			&result.Note,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.CreatedBy,
			&result.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("No se pudieron consultar los partos: %w", err)
		}
		farrowings = append(farrowings, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("No se pudieron consultar los partos: %w", err)
	}

	for _, farrowing := range farrowings {
		if farrowing.Operators, err = r.listOperators(ctx, farrowing.ID); err != nil {
			return nil, err
		}
		if farrowing.Medications, err = r.listMedications(ctx, farrowing.ID); err != nil {
			return nil, err
		}
	}
	return farrowings, nil
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
