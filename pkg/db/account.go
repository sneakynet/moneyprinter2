package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// AccountCreate creates a new account within the system.
func (db *DB) AccountCreate(a *types.Account) (uint, error) {
	res := db.d.Create(a)
	return a.ID, res.Error
}

// AccountList provides a listing of all accounts in the system.  This
// is not paginated and is one of the limiting points in the system.
func (db *DB) AccountList(filter *types.Account) ([]types.Account, error) {
	accounts := []types.Account{}
	res := db.d.Where(filter).Find(&accounts)
	return accounts, res.Error
}

// AccountGet returns a single account identified by its specific ID
func (db *DB) AccountGet(filter *types.Account) (types.Account, error) {
	acct := types.Account{}
	res := db.d.Where(filter).Preload("Premises.Wirecenter").Preload(clause.Associations).First(&acct)
	return acct, res.Error
}
