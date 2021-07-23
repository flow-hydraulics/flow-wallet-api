package handlers

import (
	"io"
	"net/http"

	gorilla "github.com/gorilla/handlers"
)

func UseCors(h http.Handler) http.Handler {
	return gorilla.CORS(gorilla.AllowedOrigins([]string{"*"}))(h)
}

func UseLogging(out io.Writer, h http.Handler) http.Handler {
	return gorilla.CombinedLoggingHandler(out, h)
}

func UseCompress(h http.Handler) http.Handler {
	return gorilla.CompressHandler(h)
}

func UseJson(h http.Handler) http.Handler {
	// Only PUT, POST, and PATCH requests are considered.
	return gorilla.ContentTypeHandler(h, "application/json")
}
