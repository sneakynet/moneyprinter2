package types

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CDR is a Call Detail Record, which provies information about a
// specific call.  This is a normalized format across all input
// formats.
type CDR struct {
	gorm.Model

	ID uint

	OrigID  uint `gorm:"unique"`
	LogTime time.Time
	CLLI    string

	CLID string `gorm:"column:clid"`
	DNIS string

	Start time.Time
	End   time.Time

	Flags uint64
}

// BillText formats the CDR so that it looks right on a bill.
func (cdr CDR) BillText() string {
	return fmt.Sprintf("Call to %s (%s)", cdr.DNIS, cdr.End.Sub(cdr.Start))
}
