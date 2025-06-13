package billing

import (
	"github.com/sneakynet/moneyprinter2/pkg/db"
	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// Processor handles some higher order actions from the Bill such as
// pre-resolving the set of fee-executors, pre-resolving all line
// items, and maintaining the database reference.
type Processor struct {
	db *db.DB

	fees map[types.FeeTarget][]Fee
}

// ProcessorOption configures the processor.
type ProcessorOption func(*Processor)

// Bill contains actualized Fees that are evaluated versions of
// database fees.
type Bill struct {
	Account types.Account
	Lines   []LineItem
}

// Fee is an interface satisfied by both static and dynamic fees.  The
// context is used to supply additional information, but need not be
// consumed if it is unneeded.
type Fee interface {
	Evaluate(FeeContext) LineItem
}

// A FeeContext is used to enable the system to calculate values for
// dynamic fees.
type FeeContext struct {
	Account types.Account
	Service types.Service
	CPE     types.NID
	CDR     types.CDR
}

// LineItem lists compose together to form bills.  These contain a
// cost as a base currency unit, which may then be localized to a
// human readable format.
type LineItem struct {
	Item string
	Fee  string
	Cost int
}
