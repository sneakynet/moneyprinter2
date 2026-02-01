package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// NIDSave persists a NID to the database.
func (db *DB) NIDSave(ctx context.Context, n *types.NID) (uint, error) {
	return n.ID, InsertOrUpdate(ctx, db.d, n)
}

// NIDList returns NIDs matching the filter.
func (db *DB) NIDList(ctx context.Context, filter *types.NID) ([]types.NID, error) {
	return gorm.G[types.NID](db.d).
		Where(filter).
		Preload("Account", nil).
		Preload("Ports", nil).
		Preload("Premise", nil).
		Preload("Premise.Wirecenter", nil).
		Find(ctx)
}

// NIDListFull returns NIDs with a truly astonishing amount of data
// loaded.
func (db *DB) NIDListFull(ctx context.Context, filter *types.NID) ([]types.NID, error) {
	return gorm.G[types.NID](db.d).
		Where(filter).
		Preload("Account", nil).
		Preload("Ports", nil).
		Preload("Ports.Services.AssignedDN", nil).
		Preload("Ports.Services.LECService", nil).
		Preload("Premise", nil).
		Preload("Premise.Wirecenter", nil).
		Find(ctx)
}

// NIDDelete removes a NID.  Use with caution.
func (db *DB) NIDDelete(ctx context.Context, n *types.NID) error {
	_, err := gorm.G[types.NID](db.d).Where(n).Delete(ctx)
	return err
}

// NIDPortSave persists a port for a NID to the database.
func (db *DB) NIDPortSave(ctx context.Context, p *types.NIDPort) (uint, error) {
	return p.ID, InsertOrUpdate(ctx, db.d, p)
}

// NIDPortAssociateService associates one or more services to a given
// NID port.  This completely replaces the association, so call with
// care.
func (db *DB) NIDPortAssociateService(_ context.Context, p *types.NIDPort, s []types.Service) error {
	return db.d.Model(p).Association("Services").Replace(s)
}
