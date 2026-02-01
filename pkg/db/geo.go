package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// WirecenterSave persists a wirecenter to the database.
func (db *DB) WirecenterSave(ctx context.Context, wc *types.Wirecenter) (uint, error) {
	return wc.ID, InsertOrUpdate(ctx, db.d, wc)
}

// WirecenterList filters the list of Wirecenters by the provided
// instance.
func (db *DB) WirecenterList(ctx context.Context, filter *types.Wirecenter) ([]types.Wirecenter, error) {
	return gorm.G[types.Wirecenter](db.d).
		Where(filter).
		Preload("Premises", nil).
		Find(ctx)
}

// WirecenterGet returns detailed information on a single Wirecenter
// selected by the parameter in the filter instance.
func (db *DB) WirecenterGet(ctx context.Context, filter *types.Wirecenter) (types.Wirecenter, error) {
	return gorm.G[types.Wirecenter](db.d).
		Where(filter).
		Preload("Premises", nil).
		First(ctx)
}

// WirecenterDelete removes a wirecenter.
func (db *DB) WirecenterDelete(ctx context.Context, wc *types.Wirecenter) error {
	_, err := gorm.G[types.Wirecenter](db.d).Where(wc).Delete(ctx)
	return err
}

// PremiseSave persists a premise to the database.
func (db *DB) PremiseSave(ctx context.Context, p *types.Premise) (uint, error) {
	return p.ID, InsertOrUpdate(ctx, db.d, p)
}

// PremiseList returns a list pf premises filtered by the provided
// instance.
func (db *DB) PremiseList(ctx context.Context, filter *types.Premise) ([]types.Premise, error) {
	return gorm.G[types.Premise](db.d).
		Where(filter).
		Preload("Account", nil).
		Preload("Wirecenter", nil).
		Find(ctx)
}

// PremiseDelete removes a premise permanently
func (db *DB) PremiseDelete(ctx context.Context, p *types.Premise) error {
	_, err := gorm.G[types.Premise](db.d).Where(p).Delete(ctx)
	return err
}
