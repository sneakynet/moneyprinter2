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

// NIDDelete removes a NID.  Use with caution.
func (db *DB) NIDDelete(n *types.NID) error {
	return db.d.Delete(n).Error
}
