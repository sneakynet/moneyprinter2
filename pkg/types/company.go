package types

import (
	"gorm.io/gorm"
)

// LEC represents a Local Exchange Company, which is an entity that
// provides one or more services.
type LEC struct {
	gorm.Model

	ID      uint
	Name    string
	Byline  string
	Contact string
	Website string
}
