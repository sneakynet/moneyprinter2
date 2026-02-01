package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// CDRSave saves a CDR into the database.  This is expected to fail if
// the OrigID is not unique, which ensures that CDRs aren't
// accidentally double-ingested.
func (db *DB) CDRSave(ctx context.Context, cdr *types.CDR) (uint, error) {
	return cdr.ID, InsertOrUpdate(ctx, db.d, cdr)
}

// CDRList returns a list of CDRs that match the given filter.
func (db *DB) CDRList(ctx context.Context, filter *types.CDR) ([]types.CDR, error) {
	return gorm.G[types.CDR](db.d).
		Where(filter).
		Find(ctx)
}
