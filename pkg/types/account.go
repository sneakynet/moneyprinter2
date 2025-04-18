package types

import (
	"gorm.io/gorm"
)

// Account represents a single entity in the system.
type Account struct {
	gorm.Model

	ID      uint
	Name    string
	Alias   string
	Contact string

	Premises []Premise
}
