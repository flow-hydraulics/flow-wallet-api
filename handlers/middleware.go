package handlers

import (
	"net/http"

	gorilla "github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func UseCors(h http.Handler) http.Handler {
	return gorilla.CORS(gorilla.AllowedOrigins([]string{"*"}))(h)
}

func UseLogging(logger *log.Entry, h http.Handler) http.Handler {
	return gorilla.CombinedLoggingHandler(logger.Logger.Out, h)
}

func UseCompress(h http.Handler) http.Handler {
	return gorilla.CompressHandler(h)
}

func UseJson(h http.Handler) http.Handler {
	// Only PUT, POST, and PATCH requests are considered.
	return gorilla.ContentTypeHandler(h, "application/json")
}
