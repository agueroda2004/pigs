package pigletfostering

import (
	"time"

	"github.com/google/uuid"
)

// PigletFosteringRole identifies how the requested sow participated in a transfer.
type PigletFosteringRole string

const (
	// PigletFosteringRoleDonor marks the requested sow as the transferring sow.
	PigletFosteringRoleDonor PigletFosteringRole = "donor"
	// PigletFosteringRoleReceiver marks the requested sow as the receiving sow.
	PigletFosteringRoleReceiver PigletFosteringRole = "receiver"
)

// PigletFosteringHistoryBySow is the read model returned for each fostering from
// the requested sow perspective. It carries only the requested sow ID, role,
// movement date, signed quantity and optional note.
type PigletFosteringHistoryBySow struct {
	SowID        uuid.UUID
	Role         PigletFosteringRole
	MovementDate time.Time
	Quantity     int
	Note         *string
}
