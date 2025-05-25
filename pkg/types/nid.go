package types

import (
	"strings"

	"gorm.io/gorm"
)

// NIDType specifies the specific type of NID that is deployed.
type NIDType uint8

const (
	// NIDTypeSRI is assigned for ShadyRate Interface NIDs (SRI).
	// This type of NID has 4 ports, supports pair remaping, and
	// supports loop-through services.
	NIDTypeSRI NIDType = iota
)

func (n NIDType) String() string {
	switch n {
	case NIDTypeSRI:
		return "SRI"
	default:
		return "UNKNOWN"
	}
}

// NID or Network Interface Device, is a device that serves as the
// point of demarcation between the LEC's facilities and the customer
// network.
type NID struct {
	gorm.Model

	ID        uint
	Account   Account
	AccountID uint
	Premise   Premise
	PremiseID uint
	CLLI      string
	Type      NIDType
	Ports     []NIDPort
}

// TableName satisfies the Tabler interface to make the name nicer.
func (n NID) TableName() string {
	return "nids"
}

// MaxPorts can be called to check how many ports the NID has
// available.
func (n NID) MaxPorts() uint8 {
	switch n.Type {
	case NIDTypeSRI:
		return 4
	default:
		return 255
	}
}

// AfterCreate sets up the NIDPorts on creation of the NID.
func (n NID) AfterCreate(tx *gorm.DB) error {
	for range n.MaxPorts() {
		tx.Save(&NIDPort{})
	}
	return nil
}

// NIDPort provides an elastic element to bind services to physical
// ports on a NID.
type NIDPort struct {
	gorm.Model

	ID              uint
	NIDID           uint
	EquipmentPortID uint
	EquipmentPort   Port

	Services []Service `gorm:"many2many:service_appearances;"`
}

// TableName satisfies the Tabler interface to have a nicer table
// name.
func (n NIDPort) TableName() string {
	return "nid_ports"
}

// ServiceList returns the names of all services assigned to the port.
func (n NIDPort) ServiceList() string {
	svcs := []string{}
	for _, svc := range n.Services {
		svcs = append(svcs, svc.LECService.Name)
	}
	return strings.Join(svcs, ",")
}

// AllDNs returns a comma separated list of DNs that can reach this
// port.
func (n NIDPort) AllDNs() string {
	dns := []string{}
	for _, svc := range n.Services {
		for _, dn := range svc.AssignedDN {
			dns = append(dns, dn.Number)
		}
	}
	return strings.Join(dns, ",")
}
