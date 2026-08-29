package types

import (
	"fmt"

	"gorm.io/gorm"
)

// Account represents a single entity in the system.
type Account struct {
	gorm.Model

	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Alias    string `json:"alias"`
	Contact  string `json:"contact"`
	BillAddr string `json:"bill_addr"`

	Premises []Premise `json:"premises"`
	Services []Service `json:"services"`
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
