package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// LECSave persists a LEC to the database.
func (db *DB) LECSave(ctx context.Context, lec *types.LEC) (uint, error) {
	return lec.ID, InsertOrUpdate(ctx, db.d, lec)
}

// LECList filters the list of LECs by the provided
// instance.
func (db *DB) LECList(ctx context.Context, filter *types.LEC) ([]types.LEC, error) {
	return gorm.G[types.LEC](db.d).
		Where(filter).
		Preload("Services", nil).
		Find(ctx)
}

// LECDelete removes a LEC.
func (db *DB) LECDelete(ctx context.Context, lec *types.LEC) error {
	_, err := gorm.G[types.LEC](db.d).Where(lec).Delete(ctx)
	return err
}

// LECServiceSave persists a LECService to the database.
func (db *DB) LECServiceSave(ctx context.Context, lec *types.LECService) (uint, error) {
	return lec.ID, InsertOrUpdate(ctx, db.d, lec)
}

// LECServiceList filters the list of LECServices by the provided
// instance.
func (db *DB) LECServiceList(ctx context.Context, filter *types.LECService) ([]types.LECService, error) {
	return gorm.G[types.LECService](db.d).
		Where(filter).
		Preload("LEC", nil).
		Find(ctx)
}

// LECServiceDelete removes a LECService.
func (db *DB) LECServiceDelete(ctx context.Context, lec *types.LECService) error {
	_, err := gorm.G[types.LECService](db.d).Where(lec).Delete(ctx)
	return err
}
