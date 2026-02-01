package web

import (
	"context"
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
	AccountSave(context.Context, *types.Account) (uint, error)
	AccountList(context.Context, *types.Account) ([]types.Account, error)
	AccountGet(context.Context, *types.Account) (types.Account, error)

	WirecenterList(context.Context, *types.Wirecenter) ([]types.Wirecenter, error)
	WirecenterGet(context.Context, *types.Wirecenter) (types.Wirecenter, error)
	WirecenterDelete(context.Context, *types.Wirecenter) error
	WirecenterSave(context.Context, *types.Wirecenter) (uint, error)

	PremiseSave(context.Context, *types.Premise) (uint, error)
	PremiseList(context.Context, *types.Premise) ([]types.Premise, error)
	PremiseDelete(context.Context, *types.Premise) error

	LECSave(context.Context, *types.LEC) (uint, error)
	LECList(context.Context, *types.LEC) ([]types.LEC, error)
	LECDelete(context.Context, *types.LEC) error

	LECServiceSave(context.Context, *types.LECService) (uint, error)
	LECServiceList(context.Context, *types.LECService) ([]types.LECService, error)
	LECServiceDelete(context.Context, *types.LECService) error

	SwitchSave(context.Context, *types.Switch) (uint, error)
	SwitchList(context.Context, *types.Switch) ([]types.Switch, error)
	SwitchDelete(context.Context, *types.Switch) error

	EquipmentSave(context.Context, *types.Equipment) (uint, error)
	EquipmentList(context.Context, *types.Equipment) ([]types.Equipment, error)
	EquipmentDelete(context.Context, *types.Equipment) error

	DNSave(context.Context, *types.DN) (uint, error)
	DNList(context.Context, *types.DN) ([]types.DN, error)
	DNListAvailable(context.Context) ([]types.DN, error)
	DNListAssigned(context.Context) ([]types.DN, error)
	DNDelete(context.Context, *types.DN) error

	PortSave(context.Context, *types.Port) (uint, error)
	PortList(context.Context, *types.Port) ([]types.Port, error)
	PortListAvailable(context.Context) ([]types.Port, error)
	PortListAssigned(context.Context) ([]types.Port, error)
	PortDelete(context.Context, *types.Port) error

	NIDSave(context.Context, *types.NID) (uint, error)
	NIDList(context.Context, *types.NID) ([]types.NID, error)
	NIDListFull(context.Context, *types.NID) ([]types.NID, error)
	NIDDelete(context.Context, *types.NID) error

	NIDPortSave(context.Context, *types.NIDPort) (uint, error)
	NIDPortAssociateService(context.Context, *types.NIDPort, []types.Service) error

	ServiceSave(context.Context, *types.Service) (uint, error)
	ServiceList(context.Context, *types.Service) ([]types.Service, error)
	ServiceListFull(context.Context, *types.Service) ([]types.Service, error)
	ServiceDelete(context.Context, *types.Service) error
	ServiceAssociateDN(context.Context, *types.Service, []types.DN) error

	FeeSave(context.Context, *types.Fee) (uint, error)
	FeeList(context.Context, *types.Fee) ([]types.Fee, error)
	FeeDelete(context.Context, *types.Fee) error

	CDRSave(context.Context, *types.CDR) (uint, error)
	CDRList(context.Context, *types.CDR) ([]types.CDR, error)

	ChargeSave(context.Context, *types.Charge) (uint, error)
	ChargeList(context.Context, *types.Charge) ([]types.Charge, error)
	ChargeDelete(context.Context, *types.Charge) error
}
