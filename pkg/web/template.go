package web

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/leekchan/accounting"

	"github.com/the-maldridge/authware"
)

//go:embed ui
var efs embed.FS

func (s *Server) templateErrorHandler(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Error while rendering template: %s\n", err)
}

func (s *Server) doTemplate(w http.ResponseWriter, r *http.Request, tmpl string, ctx pongo2.Context) {
	if ctx == nil {
		ctx = pongo2.Context{}
	}

	if user := r.Context().Value(authware.UserKey{}); user != nil {
		user = user.(authware.User)
		ctx["user"] = user
	}

	t, err := s.tpl.FromCache(tmpl)
	if err != nil {
		s.templateErrorHandler(w, err)
		return
	}
	if err := t.ExecuteWriter(ctx, w); err != nil {
		s.templateErrorHandler(w, err)
	}
}

// filterGetValueByKey gives funcrtionality that really should have
// been in the template library to begin with and allows retrieving a
// single key from a map inside the template context.
func (s *Server) filterGetValueByKey(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	m, ok := in.Interface().(map[string]interface{})
	if !ok {
		slog.Warn("Tried to convert something that isn't a map", "something", in)
		return pongo2.AsValue(nil), nil
	}
	return pongo2.AsValue(m[param.String()]), nil
}

// filterFormatMoney converts from the internal representation in the
// system of centi-cents into actual money with real units.
func (s *Server) filterFormatMoney(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	cents, ok := in.Interface().(int)
	if !ok {
		slog.Warn("Got something that wasn't a number in formatMoney", "something", in)
		return pongo2.AsValue(""), nil
	}
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	return pongo2.AsValue(ac.FormatMoney(float64(cents) / 100)), nil
}
