package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// ServiceSave persists a service.
func (db *DB) ServiceSave(s *types.Service) (uint, error) {
	res := db.d.Save(s)
	return s.ID, res.Error
}

// ServiceList retrieves services matching the filter.
func (db *DB) ServiceList(filter *types.Service) ([]types.Service, error) {
	svcs := []types.Service{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&svcs)
	return svcs, res.Error
}

// ServiceListFull retrieves services fully populated down to the
// switch and equipment.
func (db *DB) ServiceListFull(filter *types.Service) ([]types.Service, error) {
	svcs := []types.Service{}
	res := db.d.Where(filter).Preload(clause.Associations).Preload("EquipmentPort.Equipment.Switch").Find(&svcs)
	return svcs, res.Error
}

// ServiceDelete permanently deletes the matching service.
func (db *DB) ServiceDelete(s *types.Service) error {
	return db.d.Delete(s).Error
}

// ServiceAssociateDN associates one or more DNs to a given service
// entry.  This completely replaces the association, so call with
// care.
func (db *DB) ServiceAssociateDN(s *types.Service, dnList []types.DN) error {
	return db.d.Model(s).Association("AssignedDN").Replace(dnList)
}
