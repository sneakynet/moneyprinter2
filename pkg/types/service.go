package types

import (
	"strings"

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

// DNList provides a cleaner text list of the assigned DNs
func (s Service) DNList() string {
	dns := []string{}
	for _, dn := range s.AssignedDN {
		dns = append(dns, dn.Number)
	}
	return strings.Join(dns, ",")
}
