package db

import (
	"gorm.io/gorm/clause"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// LECSave persists a LEC to the database.
func (db *DB) LECSave(lec *types.LEC) (uint, error) {
	res := db.d.Save(lec)
	return lec.ID, res.Error
}

// LECList filters the list of LECs by the provided
// instance.
func (db *DB) LECList(filter *types.LEC) ([]types.LEC, error) {
	lecs := []types.LEC{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&lecs)
	return lecs, res.Error
}

// LECDelete removes a LEC.
func (db *DB) LECDelete(lec *types.LEC) error {
	res := db.d.Delete(lec)
	return res.Error
}

// LECServiceSave persists a LECService to the database.
func (db *DB) LECServiceSave(lec *types.LECService) (uint, error) {
	res := db.d.Save(lec)
	return lec.ID, res.Error
}

// LECServiceList filters the list of LECServices by the provided
// instance.
func (db *DB) LECServiceList(filter *types.LECService) ([]types.LECService, error) {
	lecs := []types.LECService{}
	res := db.d.Where(filter).Preload(clause.Associations).Find(&lecs)
	return lecs, res.Error
}

// LECServiceDelete removes a LECService.
func (db *DB) LECServiceDelete(lec *types.LECService) error {
	res := db.d.Delete(lec)
	return res.Error
}
