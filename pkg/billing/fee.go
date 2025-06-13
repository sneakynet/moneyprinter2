package billing

import (
	"log/slog"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// DynamicFee encapsulates all the logic that is used to change the
// cost based on properties in the FeeContext.
type DynamicFee struct {
	name string

	p *vm.Program
}

// NewDynamicFee configures a new dynamic fee with the given billing
// expression.
func NewDynamicFee(name, feeExpr string) (Fee, error) {
	df := DynamicFee{name: name}
	p, err := expr.Compile(feeExpr)
	if err != nil {
		return DynamicFee{}, err
	}
	df.p = p

	return df, nil
}

// Evaluate returns a LineItem based on the costs derived from the
// fee's evaluation.
func (df DynamicFee) Evaluate(fc FeeContext) LineItem {
	cost, err := expr.Run(df.p, fc)
	if err != nil {
		slog.Debug("Error evaluating fee", "fee", df.name, "context", fc)
		slog.Error("Error evaluating fee", "fee", df.name, "error", err)
		return LineItem{}
	}

	return LineItem{Fee: df.name, Cost: int(cost.(int))}
}
