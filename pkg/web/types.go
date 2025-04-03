package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/flosch/pongo2/v6"

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
}
