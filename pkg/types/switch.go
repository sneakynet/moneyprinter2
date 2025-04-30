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

// Equipment is the part of a switch that exists in a wirecenter.  It
// has ports
type Equipment struct {
	gorm.Model

	ID           uint
	SwitchID     uint
	Switch       Switch
	WirecenterID uint
	Wirecenter   Wirecenter
	Type         string
	Ports        []Port
}

// Port represents a single indivisible interface on an Equipment.
type Port struct {
	gorm.Model

	ID uint

	// Number is the port number on the equipment.
	Number      string
	Personality string
	EquipmentID uint
	Equipment   Equipment
}

// DN is a single Directory Number
type DN struct {
	gorm.Model

	ID     uint
	Number string
	CNAM   string
}
