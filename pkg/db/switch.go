package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// SwitchSave persists a switch to the database.
func (db *DB) SwitchSave(s *types.Switch) (uint, error) {
	res := db.d.Save(s)
	return s.ID, res.Error
}

// SwitchList filters the switch table and returns any hits.
func (db *DB) SwitchList(filter *types.Switch) ([]types.Switch, error) {
	switches := []types.Switch{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&switches)
	return switches, res.Error
}

// SwitchDelete removes a switch entirely.
func (db *DB) SwitchDelete(s *types.Switch) error {
	return db.d.Delete(s).Error
}
