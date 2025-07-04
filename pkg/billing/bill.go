package billing

import (
	"fmt"
	"log/slog"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// NewProcessor returns a new processor that can generate multiple
// bills.
func NewProcessor(opts ...ProcessorOption) *Processor {
	p := new(Processor)

	for _, o := range opts {
		o(p)
	}

	return p
}

// Preload fetches a bunch of information from the database so that we
// can run multiple bills without needing to re-fetch this
// information.
func (p *Processor) Preload(lec types.LEC) error {
	srcFees, err := p.db.FeeList(&types.Fee{LECReferer: lec.ID})
	if err != nil {
		slog.Error("Error preloading billing information", "error", err)
		return err
	}

	p.fees = make(map[types.FeeTarget][]Fee)

	for _, fee := range srcFees {
		newFee, err := NewDynamicFee(fee.Name, fee.Expr)
		if err != nil {
			continue
		}
		p.fees[fee.Target] = append(p.fees[fee.Target], newFee)
	}

	return nil
}

// BillAccount uses the preloaded fee information and applies that to
// an account.  WARNING: This fetches a lot of data to work out what
// services the account has and has consumed, and what it needs to be
// charged for.
func (p *Processor) BillAccount(ac types.Account, lec types.LEC) (Bill, error) {
	slog.Debug("Billing Account", "account", ac.ID, "lec", lec.ID)
	b := Bill{
		Account: ac,
		LEC:     lec,
	}

	fctx := FeeContext{Account: ac}

	// First bill anything for just having the account
	for _, fee := range p.fees[types.FeeTargetAccount] {
		l := fee.Evaluate(fctx)
		l.Item = ac.BillText()
		if l.Cost == 0 {
			continue
		}
		b.Lines = append(b.Lines, l)
	}

	// Bill for any equipment provisioned at the customer's
	// premises.  This technically gets billed for CLECs even when
	// the NID is owned by the ILEC, but presumably the CLEC is
	// being charged for use of the NID, so this is a freebie cost
	// passthrough.
	nids, err := p.db.NIDList(&types.NID{AccountID: ac.ID})
	if err != nil {
		slog.Error("Error loading NIDs for account", "account-id", ac.ID, "error", err)
		return Bill{}, err
	}
	for _, nid := range nids {
		for _, fee := range p.fees[types.FeeTargetCPE] {
			fctx.CPE = nid
			l := fee.Evaluate(fctx)
			l.Item = nid.CLLI
			if l.Cost == 0 {
				continue
			}
			b.Lines = append(b.Lines, l)
		}
	}

	// Bill for each service that is consumed.  Internal to the
	// outer loop and after the initial charge additional usage
	// based charges are assessed.
	for _, svc := range ac.Services {
		if svc.LECService.LECID != lec.ID {
			continue
		}

		for _, fee := range p.fees[types.FeeTargetService] {
			fctx.Service = svc
			l := fee.Evaluate(fctx)
			if svc.DNList() != "" {
				l.Item = fmt.Sprintf("%s (%s)", svc.LECService.Name, svc.DNList())
			} else {
				l.Item = svc.LECService.Name
			}
			if l.Cost == 0 {
				continue
			}
			b.Lines = append(b.Lines, l)
		}

		// Usage based charges go here.
		for _, dn := range svc.AssignedDN {
			slog.Debug("Billing for DN on service", "dn", dn.Number, "service", svc.LECService.Slug, "account", ac.ID)
			cdrs, err := p.db.CDRList(&types.CDR{CLID: dn.Number})
			if err != nil {
				slog.Error("Error retreiving CDRs to bill", "account", ac.ID, "dn", dn.Number, "error", err)
				return Bill{}, err
			}

			for _, cdr := range cdrs {
				for _, fee := range p.fees[types.FeeTargetUsageCDR] {
					fctx.CDR = cdr
					l := fee.Evaluate(fctx)
					l.Item = cdr.BillText()
					if l.Cost == 0 {
						continue
					}
					b.Lines = append(b.Lines, l)
				}
			}
		}
	}

	// These are the random unassigned fees that wind up on the
	// bottom of the bill.
	for _, fee := range p.fees[types.FeeTargetUnassigned] {
		l := fee.Evaluate(fctx)
		if l.Cost == 0 {
			continue
		}
		b.Lines = append(b.Lines, l)
	}

	return b, nil
}

// Cost is the total value of the entire Bill.
func (b Bill) Cost() int {
	total := 0
	for _, line := range b.Lines {
		total += line.Cost
	}
	return total
}
