package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	pigletfosteringdomain "server/internal/modules/pigletfostering/domain"
)

var (
	donorFarrowDate    = time.Date(2026, time.April, 20, 0, 0, 0, 0, time.UTC)
	receiverFarrowDate = time.Date(2026, time.April, 22, 0, 0, 0, 0, time.UTC)
	nurseWeagingDate   = time.Date(2026, time.April, 27, 0, 0, 0, 0, time.UTC)
	nowReference       = time.Date(2026, time.April, 30, 12, 0, 0, 0, time.UTC)
)

func validParams() pigletfosteringdomain.NewPigletFosteringParams {
	note := "Traslado por camada numerosa"
	return pigletfosteringdomain.NewPigletFosteringParams{
		ID:                  uuid.New(),
		DonorFarrowingID:    uuid.New(),
		ReceiverFarrowingID: uuid.New(),
		DonorSowID:          uuid.New(),
		ReceiverSowID:       uuid.New(),
		MovementDate:        time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC),
		Quantity:            2,
		Note:                &note,
		CreatedBy:           uuid.New(),
	}
}

func validBounds() pigletfosteringdomain.FosteringDateBounds {
	return pigletfosteringdomain.FosteringDateBounds{
		DonorFarrowDate:    donorFarrowDate,
		ReceiverFarrowDate: receiverFarrowDate,
	}
}

func buildFostering(t *testing.T) *pigletfosteringdomain.PigletFostering {
	t.Helper()
	fostering, err := pigletfosteringdomain.NewPigletFostering(validParams(), validBounds(), nowReference)
	if err != nil {
		t.Fatalf("NewPigletFostering() error = %v", err)
	}
	return fostering
}

func TestNewPigletFostering(t *testing.T) {
	t.Run("builds a valid fostering", func(t *testing.T) {
		params := validParams()

		fostering, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if err != nil {
			t.Fatalf("NewPigletFostering() error = %v", err)
		}
		if fostering.DonorFarrowingID != params.DonorFarrowingID || fostering.ReceiverFarrowingID != params.ReceiverFarrowingID {
			t.Fatalf("unexpected farrowings: %#v", fostering)
		}
		if fostering.Quantity != 2 || fostering.Note == nil || *fostering.Note != "Traslado por camada numerosa" {
			t.Fatalf("unexpected fields: %#v", fostering)
		}
		if !fostering.CreatedAt.Equal(nowReference) || fostering.CreatedBy != fostering.UpdatedBy {
			t.Fatalf("unexpected audit fields: %#v", fostering)
		}
	})

	t.Run("clears an empty note", func(t *testing.T) {
		params := validParams()
		empty := "   "
		params.Note = &empty

		fostering, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if err != nil {
			t.Fatalf("NewPigletFostering() error = %v", err)
		}
		if fostering.Note != nil {
			t.Fatalf("note = %v, want nil", fostering.Note)
		}
	})

	t.Run("rejects invalid required fields", func(t *testing.T) {
		for _, test := range []struct {
			name string
			set  func(*pigletfosteringdomain.NewPigletFosteringParams)
			want error
		}{
			{"id", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.ID = uuid.Nil }, pigletfosteringdomain.ErrInvalidID},
			{"donor", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.DonorFarrowingID = uuid.Nil }, pigletfosteringdomain.ErrInvalidDonor},
			{"receiver", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.ReceiverFarrowingID = uuid.Nil }, pigletfosteringdomain.ErrInvalidReceiver},
			{"donor sow", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.DonorSowID = uuid.Nil }, pigletfosteringdomain.ErrInvalidDonorSow},
			{"receiver sow", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.ReceiverSowID = uuid.Nil }, pigletfosteringdomain.ErrInvalidReceiverSow},
			{"movement date", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.MovementDate = time.Time{} }, pigletfosteringdomain.ErrInvalidMovementDate},
			{"created by", func(p *pigletfosteringdomain.NewPigletFosteringParams) { p.CreatedBy = uuid.Nil }, pigletfosteringdomain.ErrInvalidCreatedBy},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				test.set(&params)

				_, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a nil farrow date", func(t *testing.T) {
		params := validParams()
		bounds := validBounds()
		bounds.ReceiverFarrowDate = time.Time{}

		_, err := pigletfosteringdomain.NewPigletFostering(params, bounds, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects the same farrowing on both sides", func(t *testing.T) {
		params := validParams()
		params.ReceiverFarrowingID = params.DonorFarrowingID

		_, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrSameFarrowing) {
			t.Fatalf("error = %v, want ErrSameFarrowing", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validParams()
		params.Quantity = 0

		_, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("accepts the maximum quantity", func(t *testing.T) {
		params := validParams()
		params.Quantity = pigletfosteringdomain.MaxQuantity

		fostering, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if err != nil {
			t.Fatalf("NewPigletFostering() error = %v", err)
		}
		if fostering.Quantity != pigletfosteringdomain.MaxQuantity {
			t.Fatalf("quantity = %d, want %d", fostering.Quantity, pigletfosteringdomain.MaxQuantity)
		}
	})

	t.Run("rejects a quantity above the maximum", func(t *testing.T) {
		params := validParams()
		params.Quantity = pigletfosteringdomain.MaxQuantity + 1

		_, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrQuantityExceedsMax) {
			t.Fatalf("error = %v, want ErrQuantityExceedsMax", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		params := validParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("accepts a movement date on the farrow day", func(t *testing.T) {
		params := validParams()
		params.MovementDate = donorFarrowDate
		bounds := pigletfosteringdomain.FosteringDateBounds{
			DonorFarrowDate:    donorFarrowDate,
			ReceiverFarrowDate: donorFarrowDate,
		}

		fostering, err := pigletfosteringdomain.NewPigletFostering(params, bounds, nowReference)

		if err != nil {
			t.Fatalf("NewPigletFostering() error = %v", err)
		}
		if !fostering.MovementDate.Equal(donorFarrowDate) {
			t.Fatalf("movement date = %v, want %v", fostering.MovementDate, donorFarrowDate)
		}
	})

	t.Run("rejects a movement date before either farrow date", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			moving time.Time
			bounds pigletfosteringdomain.FosteringDateBounds
		}{
			{
				name:   "before both",
				moving: donorFarrowDate.AddDate(0, 0, -1),
				bounds: pigletfosteringdomain.FosteringDateBounds{
					DonorFarrowDate:    donorFarrowDate,
					ReceiverFarrowDate: donorFarrowDate,
				},
			},
			{
				name:   "before the receiver",
				moving: donorFarrowDate,
				bounds: validBounds(),
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				params.MovementDate = test.moving

				_, err := pigletfosteringdomain.NewPigletFostering(params, test.bounds, nowReference)

				if !errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforeFarrowing) {
					t.Fatalf("error = %v, want ErrMovementDateBeforeFarrowing", err)
				}
			})
		}
	})

	t.Run("accepts a movement date on a nurse partial weaging day", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			bounds func() pigletfosteringdomain.FosteringDateBounds
		}{
			{
				name: "donor on the weaging day",
				bounds: func() pigletfosteringdomain.FosteringDateBounds {
					bounds := validBounds()
					bounds.DonorPartialWeagingDate = nurseWeagingDate
					return bounds
				},
			},
			{
				name: "receiver on the weaging day",
				bounds: func() pigletfosteringdomain.FosteringDateBounds {
					bounds := validBounds()
					bounds.ReceiverPartialWeagingDate = nurseWeagingDate
					return bounds
				},
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				params.MovementDate = nurseWeagingDate

				_, err := pigletfosteringdomain.NewPigletFostering(params, test.bounds(), nowReference)

				if err != nil {
					t.Fatalf("NewPigletFostering() error = %v", err)
				}
			})
		}
	})

	t.Run("rejects a movement date before a nurse partial weaging", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			bounds func() pigletfosteringdomain.FosteringDateBounds
		}{
			{
				name: "donor before the weaging day",
				bounds: func() pigletfosteringdomain.FosteringDateBounds {
					bounds := validBounds()
					bounds.DonorPartialWeagingDate = nurseWeagingDate
					return bounds
				},
			},
			{
				name: "receiver before the weaging day",
				bounds: func() pigletfosteringdomain.FosteringDateBounds {
					bounds := validBounds()
					bounds.ReceiverPartialWeagingDate = nurseWeagingDate
					return bounds
				},
			},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				params.MovementDate = nurseWeagingDate.AddDate(0, 0, -1)

				_, err := pigletfosteringdomain.NewPigletFostering(params, test.bounds(), nowReference)

				if !errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforePartialWeaging) {
					t.Fatalf("error = %v, want ErrMovementDateBeforePartialWeaging", err)
				}
			})
		}
	})

	t.Run("accepts a movement date after a nurse partial weaging", func(t *testing.T) {
		params := validParams()
		params.MovementDate = nurseWeagingDate.AddDate(0, 0, 1)
		bounds := validBounds()
		bounds.DonorPartialWeagingDate = nurseWeagingDate

		_, err := pigletfosteringdomain.NewPigletFostering(params, bounds, nowReference)

		if err != nil {
			t.Fatalf("NewPigletFostering() error = %v", err)
		}
	})

	t.Run("rejects a movement date in the future", func(t *testing.T) {
		params := validParams()
		params.MovementDate = nowReference.AddDate(0, 0, 1)

		_, err := pigletfosteringdomain.NewPigletFostering(params, validBounds(), nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrMovementDateInFuture) {
			t.Fatalf("error = %v, want ErrMovementDateInFuture", err)
		}
	})
}

func TestPigletFosteringUpdate(t *testing.T) {
	updatedBy := uuid.New()

	t.Run("updates only the note", func(t *testing.T) {
		fostering := buildFostering(t)
		storedDate := fostering.MovementDate
		storedQuantity := fostering.Quantity
		note := "  camada ajustada  "

		err := fostering.Update(pigletfosteringdomain.UpdatePigletFosteringParams{
			Note: &note,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if !fostering.MovementDate.Equal(storedDate) || fostering.Quantity != storedQuantity {
			t.Fatalf("date or quantity changed: %#v", fostering)
		}
		if fostering.Note == nil || *fostering.Note != "camada ajustada" {
			t.Fatalf("unexpected note: %#v", fostering.Note)
		}
		if fostering.UpdatedBy != updatedBy || !fostering.UpdatedAt.Equal(nowReference) {
			t.Fatalf("unexpected audit fields: %#v", fostering)
		}
	})

	t.Run("clears the note with an empty value", func(t *testing.T) {
		fostering := buildFostering(t)
		empty := ""

		err := fostering.Update(pigletfosteringdomain.UpdatePigletFosteringParams{
			Note: &empty,
		}, updatedBy, nowReference)

		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if fostering.Note != nil {
			t.Fatalf("note = %v, want nil", fostering.Note)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		fostering := buildFostering(t)
		long := string(make([]rune, 501))

		err := fostering.Update(pigletfosteringdomain.UpdatePigletFosteringParams{
			Note: &long,
		}, updatedBy, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects an update without fields", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.Update(pigletfosteringdomain.UpdatePigletFosteringParams{}, updatedBy, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidUpdate) {
			t.Fatalf("error = %v, want ErrInvalidUpdate", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		fostering := buildFostering(t)
		note := "camada ajustada"

		err := fostering.Update(pigletfosteringdomain.UpdatePigletFosteringParams{
			Note: &note,
		}, uuid.Nil, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidUpdatedBy) {
			t.Fatalf("error = %v, want ErrInvalidUpdatedBy", err)
		}
	})

	t.Run("rejects a nil fostering", func(t *testing.T) {
		var fostering *pigletfosteringdomain.PigletFostering
		note := "camada ajustada"

		err := fostering.Update(pigletfosteringdomain.UpdatePigletFosteringParams{
			Note: &note,
		}, updatedBy, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}

func TestPigletFosteringValidateDelete(t *testing.T) {
	movementDate := time.Date(2026, time.April, 25, 0, 0, 0, 0, time.UTC)
	laterDate := movementDate.AddDate(0, 0, 1)
	deletedBy := uuid.New()

	t.Run("accepts a fostering without weagings", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{}, deletedBy)

		if err != nil {
			t.Fatalf("ValidateDelete() error = %v", err)
		}
	})

	t.Run("accepts weagings on or before the movement date", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{
			DonorWeagingDate:           movementDate.AddDate(0, 0, -1),
			DonorPartialWeagingDate:    movementDate,
			ReceiverWeagingDate:        movementDate,
			ReceiverPartialWeagingDate: movementDate.AddDate(0, 0, -2),
		}, deletedBy)

		if err != nil {
			t.Fatalf("ValidateDelete() error = %v", err)
		}
	})

	t.Run("blocks a later full weaging on the donor", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{
			DonorWeagingDate: laterDate,
		}, deletedBy)

		if !errors.Is(err, pigletfosteringdomain.ErrHasLaterWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterWeaging", err)
		}
	})

	t.Run("blocks a later full weaging on the receiver", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{
			ReceiverWeagingDate: laterDate,
		}, deletedBy)

		if !errors.Is(err, pigletfosteringdomain.ErrHasLaterWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterWeaging", err)
		}
	})

	t.Run("blocks a later partial weaging on the donor", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{
			DonorPartialWeagingDate: laterDate,
		}, deletedBy)

		if !errors.Is(err, pigletfosteringdomain.ErrHasLaterPartialWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterPartialWeaging", err)
		}
	})

	t.Run("blocks a later partial weaging on the receiver", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{
			ReceiverPartialWeagingDate: laterDate,
		}, deletedBy)

		if !errors.Is(err, pigletfosteringdomain.ErrHasLaterPartialWeaging) {
			t.Fatalf("error = %v, want ErrHasLaterPartialWeaging", err)
		}
	})

	t.Run("rejects a missing actor", func(t *testing.T) {
		fostering := buildFostering(t)

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{}, uuid.Nil)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidDeletedBy) {
			t.Fatalf("error = %v, want ErrInvalidDeletedBy", err)
		}
	})

	t.Run("rejects a nil fostering", func(t *testing.T) {
		var fostering *pigletfosteringdomain.PigletFostering

		err := fostering.ValidateDelete(pigletfosteringdomain.FosteringDeleteBounds{}, deletedBy)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidID) {
			t.Fatalf("error = %v, want ErrInvalidID", err)
		}
	})
}
