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

// ChargeSave persists a charge into the database
func (db *DB) ChargeSave(c *types.Charge) (uint, error) {
	return c.ID, db.d.Save(c).Error
}

// ChargeList returns all charges matching the given filter.
func (db *DB) ChargeList(filter *types.Charge) ([]types.Charge, error) {
	charges := []types.Charge{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&charges)
	return charges, res.Error
}

// ChargeDelete removes a charge, which may impact revenue.
func (db *DB) ChargeDelete(c *types.Charge) error {
	return db.d.Delete(c).Error
}
