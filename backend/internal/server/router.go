package server

import (
	"github.com/Ssnakerss/modmas/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	l *logger.Sl
	r *chi.Mux
}

func (s *Server) Prepare() error {

	return nil
}
