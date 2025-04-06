package types

import (
	"gorm.io/gorm"
)

// Wirecenter specifies a constrained geographic area that has all
// circuits returning to a single switching complex.
type Wirecenter struct {
	gorm.Model

	ID   uint
	Name string
}

// A Premise is a physical location within a Wirecenter.  It contains
// an address, which is a free-form string.
type Premise struct {
	gorm.Model

	ID           uint
	Address      string
	Alias        string
	Wirecenter   Wirecenter
	WirecenterID uint
}
