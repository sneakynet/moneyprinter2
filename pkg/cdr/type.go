package cdr

import (
	"io"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// Parser encapsulates all the logic that parses CDRs from files or
// remote sources.
type Parser interface {
	Parse(io.Reader, string) ([]types.CDR, error)
}
