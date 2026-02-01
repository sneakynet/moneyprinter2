package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// SwitchSave persists a switch to the database.
func (db *DB) SwitchSave(ctx context.Context, s *types.Switch) (uint, error) {
	return s.ID, InsertOrUpdate(ctx, db.d, s)
}

// SwitchList filters the switch table and returns any hits.
func (db *DB) SwitchList(ctx context.Context, filter *types.Switch) ([]types.Switch, error) {
	return gorm.G[types.Switch](db.d).
		Where(filter).
		Preload("LEC", nil).
		Find(ctx)
}

// SwitchDelete removes a switch entirely.
func (db *DB) SwitchDelete(ctx context.Context, s *types.Switch) error {
	_, err := gorm.G[types.Switch](db.d).Where(s).Delete(ctx)
	return err
}

// EquipmentSave persists an equipment to the database
func (db *DB) EquipmentSave(ctx context.Context, e *types.Equipment) (uint, error) {
	return e.ID, InsertOrUpdate(ctx, db.d, e)
}

// EquipmentList returns the equipment matching the provided filter.
func (db *DB) EquipmentList(ctx context.Context, filter *types.Equipment) ([]types.Equipment, error) {
	return gorm.G[types.Equipment](db.d).
		Where(filter).
		Preload("Switch", nil).
		Preload("Wirecenter", nil).
		Preload("Ports", nil).
		Find(ctx)
}

// EquipmentDelete removes an equipment from the system
func (db *DB) EquipmentDelete(ctx context.Context, e *types.Equipment) error {
	_, err := gorm.G[types.Equipment](db.d).Where(e).Delete(ctx)
	return err
}

// PortSave persists a port to the database
func (db *DB) PortSave(ctx context.Context, p *types.Port) (uint, error) {
	return p.ID, InsertOrUpdate(ctx, db.d, p)
}

// PortList returns the port matching the provided filter.
func (db *DB) PortList(ctx context.Context, filter *types.Port) ([]types.Port, error) {
	return gorm.G[types.Port](db.d).
		Where(filter).
		Preload("Equipment", nil).
		Find(ctx)
}

// PortListAssigned gives a list of all ports that have already been
// assigned somewhere else.
func (db *DB) PortListAssigned(ctx context.Context) ([]types.Port, error) {
	subQ := db.d.Table(types.Service{}.TableName()).Select("equipment_port_id")
	return gorm.G[types.Port](db.d).
		Where("id in (?)", subQ).
		Find(ctx)
}

// PortListAvailable gives a list of ports that are not in use
// anywhere.
func (db *DB) PortListAvailable(ctx context.Context) ([]types.Port, error) {
	subQ := db.d.Table(types.Service{}.TableName()).Select("equipment_port_id")
	return gorm.G[types.Port](db.d).
		Where("id not in (?)", subQ).
		Find(ctx)
}

// PortDelete removes an port from the system
func (db *DB) PortDelete(ctx context.Context, p *types.Port) error {
	_, err := gorm.G[types.Port](db.d).Where(p).Delete(ctx)
	return err
}

// DNSave saves a DN to the database
func (db *DB) DNSave(ctx context.Context, dn *types.DN) (uint, error) {
	return dn.ID, InsertOrUpdate(ctx, db.d, dn)
}

// DNList returns DNs matching the given filter.
func (db *DB) DNList(ctx context.Context, filter *types.DN) ([]types.DN, error) {
	return gorm.G[types.DN](db.d).
		Where(filter).
		Find(ctx)
}

// DNListAssigned filtes for DNs that are already issued in one or
// more assignments.
func (db *DB) DNListAssigned(ctx context.Context) ([]types.DN, error) {
	subQ := db.d.Table("dn_assignments").Select("dn_id")
	return gorm.G[types.DN](db.d).
		Where("id in (?)", subQ).
		Find(ctx)
}

// DNListAvailable filters for DNs that are not assigned anywhere.
func (db *DB) DNListAvailable(ctx context.Context) ([]types.DN, error) {
	subQ := db.d.Table("dn_assignments").Select("dn_id")
	return gorm.G[types.DN](db.d).
		Where("id not in (?)", subQ).
		Find(ctx)
}

// DNDelete removes a DN, use with care!
func (db *DB) DNDelete(ctx context.Context, dn *types.DN) error {
	_, err := gorm.G[types.DN](db.d).Where(dn).Delete(ctx)
	return err
}
