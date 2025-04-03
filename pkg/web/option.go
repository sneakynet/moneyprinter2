package web

// WithDB configures this server to service the specified DB
func WithDB(db DB) Option {
	return func(s *Server) { s.d = db }
}
