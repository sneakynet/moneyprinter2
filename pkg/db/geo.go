package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// WirecenterSave persists a wirecenter to the database.
func (db *DB) WirecenterSave(wc *types.Wirecenter) (uint, error) {
	res := db.d.Save(wc)
	return wc.ID, res.Error
}

// WirecenterList filters the list of Wirecenters by the provided
// instance.
func (db *DB) WirecenterList(filter *types.Wirecenter) ([]types.Wirecenter, error) {
	wcs := []types.Wirecenter{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&wcs)
	return wcs, res.Error
}

// WirecenterGet returns detailed information on a single Wirecenter
// selected by the parameter in the filter instance.
func (db *DB) WirecenterGet(filter *types.Wirecenter) (types.Wirecenter, error) {
	wc := types.Wirecenter{}
	res := db.d.Where(filter).Preload(clause.Associations).First(&wc)
	return wc, res.Error
}

// WirecenterDelete removes a wirecenter.
func (db *DB) WirecenterDelete(wc *types.Wirecenter) error {
	res := db.d.Delete(wc)
	return res.Error
}

// PremiseSave persists a premise to the database.
func (db *DB) PremiseSave(p *types.Premise) (uint, error) {
	res := db.d.Save(p)
	return p.ID, res.Error
}

// PremiseList returns a list pf premises filtered by the provided
// instance.
func (db *DB) PremiseList(filter *types.Premise) ([]types.Premise, error) {
	premises := []types.Premise{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&premises)
	return premises, res.Error
}

// PremiseDelete removes a premise permanently
func (db *DB) PremiseDelete(p *types.Premise) error {
	return db.d.Delete(p).Error
}
