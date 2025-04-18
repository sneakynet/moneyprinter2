package types

import (
	"gorm.io/gorm"
)

// Wirecenter specifies a constrained geographic area that has all
// circuits returning to a single switching complex.
type Wirecenter struct {
	gorm.Model

	ID       uint
	Name     string
	Premises []Premise
}

// A Premise is a physical location within a Wirecenter.  It contains
// an address, which is a free-form string.
type Premise struct {
	gorm.Model

	ID           uint
	AccountID    uint
	Account      Account
	Address      string
	Wirecenter   Wirecenter
	WirecenterID uint
}
