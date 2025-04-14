package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// Server handles the HTTP frontend.
type Server struct {
	r chi.Router
	n *http.Server
	d DB

	tpl *pongo2.TemplateSet
}

// Option configures the Server
type Option func(*Server)

// DB sets the specific calls that the web layer will make.
type DB interface {
	AccountCreate(*types.Account) (uint, error)
	AccountList(*types.Account) ([]types.Account, error)
	AccountGet(*types.Account) (types.Account, error)

	WirecenterList(*types.Wirecenter) ([]types.Wirecenter, error)
	WirecenterGet(*types.Wirecenter) (types.Wirecenter, error)
	WirecenterDelete(*types.Wirecenter) error
	WirecenterSave(*types.Wirecenter) (uint, error)

	PremiseSave(*types.Premise) (uint, error)
	PremiseList(*types.Premise) ([]types.Premise, error)
	PremiseDelete(*types.Premise) error

	LECSave(*types.LEC) (uint, error)
	LECList(*types.LEC) ([]types.LEC, error)
	LECDelete(*types.LEC) error

	SwitchSave(*types.Switch) (uint, error)
	SwitchList(*types.Switch) ([]types.Switch, error)
	SwitchDelete(*types.Switch) error

	EquipmentSave(*types.Equipment) (uint, error)
	EquipmentList(*types.Equipment) ([]types.Equipment, error)
	EquipmentDelete(*types.Equipment) error

	PortSave(*types.Port) (uint, error)
	PortList(*types.Port) ([]types.Port, error)
	PortDelete(*types.Port) error
}
