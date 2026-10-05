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

func TestNewPigletFostering(t *testing.T) {
	t.Run("builds a valid fostering", func(t *testing.T) {
		params := validParams()

		fostering, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

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

		fostering, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

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

				_, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

				if !errors.Is(err, test.want) {
					t.Fatalf("error = %v, want %v", err, test.want)
				}
			})
		}
	})

	t.Run("rejects a nil farrow date", func(t *testing.T) {
		params := validParams()

		_, err := pigletfosteringdomain.NewPigletFostering(params, time.Time{}, receiverFarrowDate, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidFarrowDate) {
			t.Fatalf("error = %v, want ErrInvalidFarrowDate", err)
		}
	})

	t.Run("rejects the same farrowing on both sides", func(t *testing.T) {
		params := validParams()
		params.ReceiverFarrowingID = params.DonorFarrowingID

		_, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrSameFarrowing) {
			t.Fatalf("error = %v, want ErrSameFarrowing", err)
		}
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		params := validParams()
		params.Quantity = 0

		_, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidQuantity) {
			t.Fatalf("error = %v, want ErrInvalidQuantity", err)
		}
	})

	t.Run("rejects a note that is too long", func(t *testing.T) {
		params := validParams()
		long := string(make([]rune, 501))
		params.Note = &long

		_, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrInvalidNote) {
			t.Fatalf("error = %v, want ErrInvalidNote", err)
		}
	})

	t.Run("rejects a movement date on or before either farrow date", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			moving time.Time
		}{
			{"equal to donor", donorFarrowDate},
			{"equal to receiver", receiverFarrowDate},
			{"before both", donorFarrowDate.AddDate(0, 0, -1)},
		} {
			t.Run(test.name, func(t *testing.T) {
				params := validParams()
				params.MovementDate = test.moving

				_, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

				if !errors.Is(err, pigletfosteringdomain.ErrMovementDateBeforeFarrowing) {
					t.Fatalf("error = %v, want ErrMovementDateBeforeFarrowing", err)
				}
			})
		}
	})

	t.Run("rejects a movement date in the future", func(t *testing.T) {
		params := validParams()
		params.MovementDate = nowReference.AddDate(0, 0, 1)

		_, err := pigletfosteringdomain.NewPigletFostering(params, donorFarrowDate, receiverFarrowDate, nowReference)

		if !errors.Is(err, pigletfosteringdomain.ErrMovementDateInFuture) {
			t.Fatalf("error = %v, want ErrMovementDateInFuture", err)
		}
	})
}
