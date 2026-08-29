package types

import (
	"gorm.io/gorm"
)

// LEC represents a Local Exchange Company, which is an entity that
// provides one or more services.
type LEC struct {
	gorm.Model

	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Byline  string `json:"byline"`
	Contact string `json:"contact"`
	Website string `json:"website"`

	Services []LECService `json:"services"`
}

// LECService defines a service that is provided by a LEC.
type LECService struct {
	gorm.Model

	ID          uint `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	LECID       uint `json:"lec_id"`
	LEC         LEC `json:"lec"`
}
