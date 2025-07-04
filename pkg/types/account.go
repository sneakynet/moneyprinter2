package types

import (
	"fmt"

	"gorm.io/gorm"
)

// Account represents a single entity in the system.
type Account struct {
	gorm.Model

	ID       uint
	Name     string
	Alias    string
	Contact  string
	BillAddr string

	Premises []Premise
	Services []Service
}

// BillText formats the account number for printing on a bill.
func (a Account) BillText() string {
	return fmt.Sprintf("Account #%d", a.ID)
}

// LECList returns a list of all unique LECs that this account is
// doing business with.
func (a Account) LECList() []LEC {
	tmp := make(map[uint]LEC)
	for _, s := range a.Services {
		if _, ok := tmp[s.LECService.LECID]; ok {
			continue
		}
		tmp[s.LECService.LECID] = s.LECService.LEC
	}
	out := []LEC{}
	for _, lec := range tmp {
		out = append(out, lec)
	}
	return out
}
