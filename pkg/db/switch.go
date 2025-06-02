package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// SwitchSave persists a switch to the database.
func (db *DB) SwitchSave(s *types.Switch) (uint, error) {
	res := db.d.Save(s)
	return s.ID, res.Error
}

// SwitchList filters the switch table and returns any hits.
func (db *DB) SwitchList(filter *types.Switch) ([]types.Switch, error) {
	switches := []types.Switch{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&switches)
	return switches, res.Error
}

// SwitchDelete removes a switch entirely.
func (db *DB) SwitchDelete(s *types.Switch) error {
	return db.d.Delete(s).Error
}

// EquipmentSave persists an equipment to the database
func (db *DB) EquipmentSave(e *types.Equipment) (uint, error) {
	res := db.d.Save(e)
	return e.ID, res.Error
}

// EquipmentList returns the equipment matching the provided filter.
func (db *DB) EquipmentList(filter *types.Equipment) ([]types.Equipment, error) {
	equipment := []types.Equipment{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&equipment)
	return equipment, res.Error
}

// EquipmentDelete removes an equipment from the system
func (db *DB) EquipmentDelete(e *types.Equipment) error {
	return db.d.Delete(e).Error
}

// PortSave persists a port to the database
func (db *DB) PortSave(p *types.Port) (uint, error) {
	res := db.d.Save(p)
	return p.ID, res.Error
}

// PortList returns the port matching the provided filter.
func (db *DB) PortList(filter *types.Port) ([]types.Port, error) {
	ports := []types.Port{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&ports)
	return ports, res.Error
}

// PortListAssigned gives a list of all ports that have already been
// assigned somewhere else.
func (db *DB) PortListAssigned() ([]types.Port, error) {
	ports := []types.Port{}
	res := db.d.Where("id in (?)", db.d.Table(types.NIDPort{}.TableName()).Select("equipment_port_id")).Find(&ports)
	return ports, res.Error
}

// PortListAvailable gives a list of ports that are not in use
// anywhere.
func (db *DB) PortListAvailable() ([]types.Port, error) {
	ports := []types.Port{}
	res := db.d.Where("id not in (?)", db.d.Table(types.NIDPort{}.TableName()).Select("equipment_port_id")).Find(&ports)
	return ports, res.Error
}

// PortDelete removes an port from the system
func (db *DB) PortDelete(p *types.Port) error {
	return db.d.Delete(p).Error
}

// DNSave saves a DN to the database
func (db *DB) DNSave(dn *types.DN) (uint, error) {
	res := db.d.Save(dn)
	return dn.ID, res.Error
}

// DNList returns DNs matching the given filter.
func (db *DB) DNList(filter *types.DN) ([]types.DN, error) {
	DNList := []types.DN{}
	res := db.d.Where(filter).Find(&DNList)
	return DNList, res.Error
}

// DNListAssigned filtes for DNs that are already issued in one or
// more assignments.
func (db *DB) DNListAssigned() ([]types.DN, error) {
	DNList := []types.DN{}
	res := db.d.Where("id in (?)", db.d.Table("dn_assignments").Select("dn_id")).Find(&DNList)
	return DNList, res.Error
}

// DNListAvailable filters for DNs that are not assigned anywhere.
func (db *DB) DNListAvailable() ([]types.DN, error) {
	DNList := []types.DN{}
	res := db.d.Where("id not in (?)", db.d.Table("dn_assignments").Select("dn_id")).Find(&DNList)
	return DNList, res.Error
}

// DNDelete removes a DN, use with care!
func (db *DB) DNDelete(dn *types.DN) error {
	return db.d.Delete(dn).Error
}
