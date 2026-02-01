package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// FeeSave persists a fee into the database
func (db *DB) FeeSave(ctx context.Context, f *types.Fee) (uint, error) {
	return f.ID, InsertOrUpdate(ctx, db.d, f)
}

// FeeList returns all fees that match the filter.
func (db *DB) FeeList(ctx context.Context, filter *types.Fee) ([]types.Fee, error) {
	return gorm.G[types.Fee](db.d).
		Preload("AssessedBy", nil).
		Find(ctx)
}

// FeeDelete removes the fee from the database.
func (db *DB) FeeDelete(ctx context.Context, f *types.Fee) error {
	_, err := gorm.G[types.Fee](db.d).
		Where(f).
		Delete(ctx)
	return err
}

// ChargeSave persists a charge into the database
func (db *DB) ChargeSave(ctx context.Context, c *types.Charge) (uint, error) {
	return c.ID, InsertOrUpdate(ctx, db.d, c)
}

// ChargeList returns all charges matching the given filter.
func (db *DB) ChargeList(ctx context.Context, filter *types.Charge) ([]types.Charge, error) {
	return gorm.G[types.Charge](db.d).
		Where(filter).
		Preload("AssessedBy", nil).
		Find(ctx)
}

// ChargeDelete removes a charge, which may impact revenue.
func (db *DB) ChargeDelete(ctx context.Context, c *types.Charge) error {
	_, err := gorm.G[types.Charge](db.d).Where(c).Delete(ctx)
	return err
}
