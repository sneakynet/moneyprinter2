package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// FeeSave persists a fee into the database
func (db *DB) FeeSave(f *types.Fee) (uint, error) {
	res := db.d.Save(f)
	return f.ID, res.Error
}

// FeeList returns all fees that match the filter.
func (db *DB) FeeList(filter *types.Fee) ([]types.Fee, error) {
	fees := []types.Fee{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&fees)
	return fees, res.Error
}

// FeeDelete removes the fee from the database.
func (db *DB) FeeDelete(f *types.Fee) error {
	return db.d.Delete(f).Error
}
