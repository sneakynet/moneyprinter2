package types

import (
	"gorm.io/gorm"
)

// Service is an instantiation of a service that people want to order.
type Service struct {
	gorm.Model

	ID           uint
	LECServiceID uint
	LECService   LECService
	AccountID    uint
	Account      Account

	AssignedDN []DN `gorm:"many2many:dn_assignments;"`
}
