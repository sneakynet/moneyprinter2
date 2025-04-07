package types

import (
	"gorm.io/gorm"
)

// Switch represents the top level physical device that switches
// elements of the network.  The switch does not directly present an
// interface towards a customer.  These interfaces are presented to a
// customer on a NID, which via Outside Plant (OSP) is connected back
// to one or more line equipment devices.
type Switch struct {
	gorm.Model

	ID    uint
	CLLI  string
	Alias string
	LECID uint
	LEC   LEC
}
