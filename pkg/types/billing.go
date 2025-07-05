package types

import (
	"gorm.io/gorm"
)

// FeeTarget defines what a given Fee applies to, be it a service or
// account or usage based item.
type FeeTarget uint

const (
	// FeeTargetUnassigned is a fee that is assessed once per bill
	// as an assorted fee with no other logic.
	FeeTargetUnassigned FeeTarget = iota

	// FeeTargetAccount acts on an entire account.
	FeeTargetAccount

	// FeeTargetService acts on services that are instantiated and
	// ordered, for example the cost of provisioning a line.
	FeeTargetService

	// FeeTargetCPE acts on equipment that has been provisioned to
	// support a customer's service at their premises.
	FeeTargetCPE

	// FeeTargetUsageCDR acts on usage of a service, as determined
	// by a CDR (telephony).
	FeeTargetUsageCDR
)

func (ft FeeTarget) String() string {
	switch ft {
	case FeeTargetUnassigned:
		return "Unassigned / General"
	case FeeTargetAccount:
		return "Account"
	case FeeTargetService:
		return "Service"
	case FeeTargetCPE:
		return "CPE"
	case FeeTargetUsageCDR:
		return "Usage - CDR"
	default:
		return "UNKNOWN FEE TARGET"
	}
}

// A Fee is an individual line item that comprises a bill.  A bill is
// composed of fees as calulated for an account.  Fees can match
// against many different facets of an account and are evaluated
// within a FeeContext which includes details about an account's
// services and associated infrastructure.
type Fee struct {
	gorm.Model

	ID   uint
	Name string
	Expr string

	Target FeeTarget

	AssessedBy LEC `gorm:"foreignKey:LECReferer"`
	LECReferer uint
}

// A Charge directly resolves to a LineItem and is a pass through type
// to enable making one-off charges to an account.
type Charge struct {
	ID uint

	AccountID  uint
	AssessedBy LEC `gorm:"foreignKey:LECReferer"`
	LECReferer uint

	Item string
	Cost int
}
