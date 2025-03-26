package types

import (
	"gorm.io/gorm"
)

// A LEC or Local Exchange Company is the base unit of multi-tenancy
// in moneyprinter.  The LEC is the operating company that services
// customers.
type LEC struct {
	gorm.Model

	ID uint
	Name string
	Byline string
	Contact string
	Website string
}

// Service defines one flat-rate billed services that a LEC provides
// and an account can have.
type Service struct {
	gorm.Model

	ID uint
	Name string
	Slug string
}

// An Account is the base unit of a customer bill.  This is the
// element that unifies all of a customer's services, their bills, and
// any credentials they might have that moneyprinter knows about.
type Account struct {
	gorm.Model

	ID uint
	Name string
	Alias string
	Contact string
}

// Premise refers to a location where service may be delivered.  The
// intention is that these function somewhat like addresses, and allow
// for quick identification of where things are.
type Premise struct {
	gorm.Model

	ID uint
	Address string
}

// A Wirecenter represents a constrained physical area that is served
// from a single collection of equipment.
type Wirecenter struct {
	gorm.Model

	ID uint
	Name string
}
