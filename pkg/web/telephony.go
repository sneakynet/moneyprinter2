package web

import (
	"fmt"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

//////////////
// Switches //
//////////////

func (s *Server) uiViewSwitchList(w http.ResponseWriter, r *http.Request) {
	switches, err := s.d.SwitchList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/switch/list.p2", pongo2.Context{"switches": switches})
}

func (s *Server) uiViewSwitchDetail(w http.ResponseWriter, r *http.Request) {
	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/switch/detail.p2", pongo2.Context{"switch": switches[0]})
}

func (s *Server) uiViewSwitchFormSingle(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/switch/form_single.p2", pongo2.Context{"lecs": lecs})
}

func (s *Server) uiViewSwitchEdit(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"switch": switches[0],
		"lecs":   lecs,
	}

	s.doTemplate(w, r, "views/switch/form_single.p2", ctx)
}

func (s *Server) uiViewSwitchUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	sw := types.Switch{
		ID:    s.strToUint(chi.URLParam(r, "id")),
		CLLI:  r.FormValue("switch_clli"),
		Alias: r.FormValue("switch_alias"),
		LECID: s.strToUint(r.FormValue("switch_lec")),
	}

	id, err := s.d.SwitchSave(&sw)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%d", id), http.StatusSeeOther)
}

func (s *Server) uiViewSwitchDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.SwitchDelete(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, "/ui/admin/switches", http.StatusSeeOther)
}

///////////////
// Equipment //
///////////////

func (s *Server) uiViewEquipmentList(w http.ResponseWriter, r *http.Request) {
	filter := &types.Equipment{}

	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	if len(switches) > 0 {
		filter.SwitchID = switches[0].ID
	}

	equipment, err := s.d.EquipmentList(filter)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"switches":  switches,
		"equipment": equipment,
	}
	s.doTemplate(w, r, "views/equipment/list.p2", ctx)
}

func (s *Server) uiViewEquipmentDetail(w http.ResponseWriter, r *http.Request) {
	equipment, err := s.d.EquipmentList(&types.Equipment{ID: s.strToUint(chi.URLParam(r, "eid"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/equipment/detail.p2", pongo2.Context{"equipment": equipment[0]})
}

func (s *Server) uiViewEquipmentEdit(w http.ResponseWriter, r *http.Request) {
	equipment, err := s.d.EquipmentList(&types.Equipment{ID: s.strToUint(chi.URLParam(r, "eid"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	wirecenters, err := s.d.WirecenterList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx := pongo2.Context{
		"switches":    switches,
		"wirecenters": wirecenters,
		"equipment":   equipment[0],
	}
	s.doTemplate(w, r, "views/equipment/form_single.p2", ctx)
}

func (s *Server) uiViewEquipmentFormSingle(w http.ResponseWriter, r *http.Request) {
	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	wirecenters, err := s.d.WirecenterList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx := pongo2.Context{
		"switches":    switches,
		"wirecenters": wirecenters,
	}

	s.doTemplate(w, r, "views/equipment/form_single.p2", ctx)
}

func (s *Server) uiViewEquipmentUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	equipment := types.Equipment{
		ID:           s.strToUint(chi.URLParam(r, "eid")),
		SwitchID:     s.strToUint(r.FormValue("equipment_switch")),
		WirecenterID: s.strToUint(r.FormValue("equipment_wirecenter")),
		Type:         r.FormValue("equipment_type"),
	}

	_, err := s.d.EquipmentSave(&equipment)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%d/equipment", equipment.SwitchID), http.StatusSeeOther)
}

func (s *Server) uiViewEquipmentDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.EquipmentDelete(&types.Equipment{ID: s.strToUint(chi.URLParam(r, "eid"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%s/equipment", chi.URLParam(r, "id")), http.StatusSeeOther)
}

///////////
// Ports //
///////////

func (s *Server) uiViewPortList(w http.ResponseWriter, r *http.Request) {
	equipment, err := s.d.EquipmentList(&types.Equipment{ID: s.strToUint(chi.URLParam(r, "eid"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ports, err := s.d.PortList(&types.Port{EquipmentID: equipment[0].ID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx := pongo2.Context{
		"equipment": equipment[0],
		"ports":     ports,
	}

	s.doTemplate(w, r, "views/port/list.p2", ctx)
}

func (s *Server) uiViewPortEdit(w http.ResponseWriter, r *http.Request) {
	ports, err := s.d.PortList(&types.Port{ID: s.strToUint(chi.URLParam(r, "pid"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/port/form_single.p2", pongo2.Context{"port": ports[0]})
}

func (s *Server) uiViewPortUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	port := types.Port{
		ID:          s.strToUint(chi.URLParam(r, "pid")),
		Number:      r.FormValue("port_number"),
		Personality: r.FormValue("port_personality"),
		EquipmentID: s.strToUint(chi.URLParam(r, "eid")),
	}

	if _, err := s.d.PortSave(&port); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%s/equipment/%s/ports", chi.URLParam(r, "id"), chi.URLParam(r, "eid")), http.StatusSeeOther)
}

func (s *Server) uiViewPortFormSingle(w http.ResponseWriter, r *http.Request) {
	equipment, err := s.d.EquipmentList(&types.Equipment{ID: s.strToUint(chi.URLParam(r, "eid"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/port/form_single.p2", pongo2.Context{"equipment": equipment[0]})
}

func (s *Server) uiViewPortFormBulk(w http.ResponseWriter, r *http.Request) {
	equipment, err := s.d.EquipmentList(&types.Equipment{ID: s.strToUint(chi.URLParam(r, "eid"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/port/form_bulk.p2", pongo2.Context{"equipment": equipment[0]})
}

func (s *Server) uiViewPortFormBulkSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	tpl, err := pongo2.FromString(r.FormValue("port_tmpl"))
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	for id := range s.strToUint(r.FormValue("port_count")) {
		number, err := tpl.Execute(pongo2.Context{"id": id})
		if err != nil {
			s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
			return
		}

		port := types.Port{
			Number:      number,
			Personality: r.FormValue("port_personality"),
			EquipmentID: s.strToUint(chi.URLParam(r, "eid")),
		}

		if _, err := s.d.PortSave(&port); err != nil {
			s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
			return
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%s/equipment/%s/ports", chi.URLParam(r, "id"), chi.URLParam(r, "eid")), http.StatusSeeOther)
}

func (s *Server) uiViewPortDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.PortDelete(&types.Port{ID: s.strToUint(chi.URLParam(r, "pid"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%s/equipment/%s/ports/", chi.URLParam(r, "id"), chi.URLParam(r, "eid")), http.StatusSeeOther)
}
