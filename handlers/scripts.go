package handlers

import (
	"log"
	"net/http"

	"github.com/eqlabs/flow-wallet-service/scripts"
)

type Scripts struct {
	log     *log.Logger
	service *scripts.Service
}

// NewScripts initiates a new scripts server.
func NewScripts(l *log.Logger, s *scripts.Service) *Scripts {
	return &Scripts{l, s}
}

func (s *Scripts) Execute() http.Handler {
	h := http.HandlerFunc(s.ExecuteFunc)
	return UseJson(h)
}
