package tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/handlers"
	"github.com/gorilla/mux"
)

func Test_IdempotencyMiddleware(t *testing.T) {
	is := handlers.NewIdempotencyStoreLocal()

	// Dummy endpoint for testing
	testHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	router := mux.NewRouter()
	router.Handle("/test", handlers.UseIdempotency(testHandler, handlers.IdempotencyHandlerOptions{
		Expiry:      5000 * time.Millisecond,
		IgnorePaths: []string{"/ignored"},
	}, is)).Methods(http.MethodPost)

	ik := "idempotency-key-test"
	body := bytes.NewBufferString("")

	t.Run("returns 200 with a fresh key", func(t *testing.T) {
		res := sendWithHeaders(router, http.MethodPost, "/test", body, map[string]string{"Idempotency-Key": ik})
		assertStatusCode(t, res, http.StatusOK)
	})

	t.Run("returns 409 with a used key", func(t *testing.T) {
		res := sendWithHeaders(router, http.MethodPost, "/test", body, map[string]string{"Idempotency-Key": ik})
		assertStatusCode(t, res, http.StatusConflict)
	})

	t.Run("returns 400 with missing header", func(t *testing.T) {
		res := send(router, http.MethodPost, "/test", body)
		assertStatusCode(t, res, http.StatusBadRequest)
	})

}

// TODO: Move to test utils
func sendWithHeaders(router *mux.Router, method, path string, body io.Reader, headers map[string]string) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("content-type", "application/json")

	for hk, hv := range headers {
		req.Header.Set(hk, hv)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Result()
}
