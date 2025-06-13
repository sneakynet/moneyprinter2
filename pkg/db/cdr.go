package db

import (
	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// CDRSave saves a CDR into the database.  This is expected to fail if
// the OrigID is not unique, which ensures that CDRs aren't
// accidentally double-ingested.
func (db *DB) CDRSave(cdr *types.CDR) (uint, error) {
	res := db.d.Save(cdr)
	return cdr.ID, res.Error
}

// CDRList returns a list of CDRs that match the given filter.
func (db *DB) CDRList(filter *types.CDR) ([]types.CDR, error) {
	cdrs := []types.CDR{}
	res := db.d.Where(filter).Find(&cdrs)
	return cdrs, res.Error
}
