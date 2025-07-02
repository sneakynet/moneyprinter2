package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// NIDSave persists a NID to the database.
func (db *DB) NIDSave(n *types.NID) (uint, error) {
	res := db.d.Save(n)
	return n.ID, res.Error
}

// NIDList returns NIDs matching the filter.
func (db *DB) NIDList(filter *types.NID) ([]types.NID, error) {
	NIDs := []types.NID{}
	res := db.d.Where(filter).Preload(clause.Associations).Preload("Premise.Wirecenter").Find(&NIDs)
	return NIDs, res.Error
}

// NIDListFull returns NIDs with a truly astonishing amount of data
// loaded.
func (db *DB) NIDListFull(filter *types.NID) ([]types.NID, error) {
	NIDs := []types.NID{}
	res := db.d.Where(filter).Preload(clause.Associations).Preload("Premise.Wirecenter").Preload("Ports.Services.LECService").Preload("Ports.Services.AssignedDN").Find(&NIDs)
	return NIDs, res.Error
}

// NIDDelete removes a NID.  Use with caution.
func (db *DB) NIDDelete(n *types.NID) error {
	return db.d.Delete(n).Error
}

// NIDPortSave persists a port for a NID to the database.
func (db *DB) NIDPortSave(p *types.NIDPort) (uint, error) {
	res := db.d.Save(p)
	return p.ID, res.Error
}

// NIDPortAssociateService associates one or more services to a given
// NID port.  This completely replaces the association, so call with
// care.
func (db *DB) NIDPortAssociateService(p *types.NIDPort, s []types.Service) error {
	return db.d.Model(p).Association("Services").Replace(s)
}
