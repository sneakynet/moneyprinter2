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
	AccountSave(*types.Account) (uint, error)
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

	LECServiceSave(*types.LECService) (uint, error)
	LECServiceList(*types.LECService) ([]types.LECService, error)
	LECServiceDelete(*types.LECService) error

	SwitchSave(*types.Switch) (uint, error)
	SwitchList(*types.Switch) ([]types.Switch, error)
	SwitchDelete(*types.Switch) error

	EquipmentSave(*types.Equipment) (uint, error)
	EquipmentList(*types.Equipment) ([]types.Equipment, error)
	EquipmentDelete(*types.Equipment) error

	DNSave(*types.DN) (uint, error)
	DNList(*types.DN) ([]types.DN, error)
	DNListAvailable() ([]types.DN, error)
	DNListAssigned() ([]types.DN, error)
	DNDelete(*types.DN) error

	PortSave(*types.Port) (uint, error)
	PortList(*types.Port) ([]types.Port, error)
	PortListAvailable() ([]types.Port, error)
	PortListAssigned() ([]types.Port, error)
	PortDelete(*types.Port) error

	NIDSave(*types.NID) (uint, error)
	NIDList(*types.NID) ([]types.NID, error)
	NIDListFull(*types.NID) ([]types.NID, error)
	NIDDelete(*types.NID) error

	NIDPortSave(*types.NIDPort) (uint, error)
	NIDPortAssociateService(*types.NIDPort, []types.Service) error

	ServiceSave(*types.Service) (uint, error)
	ServiceList(*types.Service) ([]types.Service, error)
	ServiceListFull(*types.Service) ([]types.Service, error)
	ServiceDelete(*types.Service) error
	ServiceAssociateDN(*types.Service, []types.DN) error

	FeeSave(*types.Fee) (uint, error)
	FeeList(*types.Fee) ([]types.Fee, error)
	FeeDelete(*types.Fee) error

	CDRSave(*types.CDR) (uint, error)
	CDRList(*types.CDR) ([]types.CDR, error)

	ChargeSave(*types.Charge) (uint, error)
	ChargeList(*types.Charge) ([]types.Charge, error)
	ChargeDelete(*types.Charge) error
}
