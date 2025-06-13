package types

import (
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

	CLID uint `gorm:"column:clid"`
	DNIS uint

	Start time.Time
	End   time.Time

	Flags uint64
}
