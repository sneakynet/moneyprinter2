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
