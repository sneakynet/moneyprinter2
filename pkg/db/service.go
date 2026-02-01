package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// ServiceSave persists a service.
func (db *DB) ServiceSave(ctx context.Context, s *types.Service) (uint, error) {
	return s.ID, InsertOrUpdate(ctx, db.d, s)
}

// ServiceList retrieves services matching the filter.
func (db *DB) ServiceList(ctx context.Context, filter *types.Service) ([]types.Service, error) {
	return gorm.G[types.Service](db.d).
		Where(filter).
		Preload("Account", nil).
		Preload("LECService", nil).
		Preload("EquipmentPort", nil).
		Preload("AssignedDN", nil).
		Find(ctx)
}

// ServiceListFull retrieves services fully populated down to the
// switch and equipment.
func (db *DB) ServiceListFull(ctx context.Context, filter *types.Service) ([]types.Service, error) {
	return gorm.G[types.Service](db.d).
		Where(filter).
		Preload("Account", nil).
		Preload("LECService", nil).
		Preload("EquipmentPort", nil).
		Preload("EquipmentPort.Equipment.Switch", nil).
		Preload("AssignedDN", nil).
		Find(ctx)
}

// ServiceDelete permanently deletes the matching service.
func (db *DB) ServiceDelete(ctx context.Context, s *types.Service) error {
	_, err := gorm.G[types.Service](db.d).Where(s).Delete(ctx)
	return err
}

// ServiceAssociateDN associates one or more DNs to a given service
// entry.  This completely replaces the association, so call with
// care.
func (db *DB) ServiceAssociateDN(_ context.Context, s *types.Service, dnList []types.DN) error {
	return db.d.Model(s).Association("AssignedDN").Replace(dnList)
}
