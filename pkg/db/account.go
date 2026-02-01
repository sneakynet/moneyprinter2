package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// AccountSave creates a new account within the system.
func (db *DB) AccountSave(ctx context.Context, a *types.Account) (uint, error) {
	return a.ID, InsertOrUpdate(ctx, db.d, a)
}

// AccountList provides a listing of all accounts in the system.  This
// is not paginated and is one of the limiting points in the system.
func (db *DB) AccountList(ctx context.Context, filter *types.Account) ([]types.Account, error) {
	accounts, err := gorm.G[types.Account](db.d).
		Where(filter).
		Find(ctx)
	return accounts, err
}

// AccountGet returns a single account identified by its specific ID
func (db *DB) AccountGet(ctx context.Context, filter *types.Account) (types.Account, error) {
	acct, err := gorm.G[types.Account](db.d).
		Where(filter).
		Preload("Premises", nil).
		Preload("Services", nil).
		Preload("Premises.Wirecenter", nil).
		Preload("Services.LECService", nil).
		Preload("Services.LECService.LEC", nil).
		Preload("Services.AssignedDN", nil).
		Preload("Services.EquipmentPort", nil).
		First(ctx)
	return acct, err
}
