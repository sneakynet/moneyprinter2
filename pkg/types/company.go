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

	Services []LECService
}

// LECService defines a service that is provided by a LEC.
type LECService struct {
	gorm.Model

	ID          uint
	Name        string
	Slug        string
	Description string
	LECID       uint
	LEC         LEC
}
