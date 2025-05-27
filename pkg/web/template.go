package web

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/flosch/pongo2/v6"
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
