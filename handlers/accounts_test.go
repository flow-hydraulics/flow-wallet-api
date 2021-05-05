package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/eqlabs/flow-wallet-service/account"
	"github.com/eqlabs/flow-wallet-service/data"
	"github.com/gorilla/mux"
)

func TestAccountHandlers(t *testing.T) {
	logger := log.New(ioutil.Discard, "", log.LstdFlags|log.Lshortfile)

	service, err := account.SetupTestService(logger)
	if err != nil {
		t.Fatalf("Error while running setup: %s", err)
	}

	var tempAcc data.Account

	handlers := NewAccounts(logger, service)

	router := mux.NewRouter()
	router.HandleFunc("/", handlers.List).Methods("GET")
	router.HandleFunc("/", handlers.Create).Methods("POST")
	router.HandleFunc("/{address}", handlers.Details).Methods("GET")

	// NOTE: The order of the test "steps" matters
	steps := []struct {
		name     string
		method   string
		url      string
		expected string
		status   int
	}{
		{
			name:     "HTTP GET accounts.List db empty",
			method:   "GET",
			url:      "/",
			expected: `\[\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP POST accounts.Create",
			method:   "POST",
			url:      "/",
			expected: `\{"address":".*","createdAt":".*","updatedAt":".*"\}\n`,
			status:   http.StatusCreated,
		},
		{
			name:     "HTTP GET accounts.List db not empty",
			method:   "GET",
			url:      "/",
			expected: `\[\{"address":".*","createdAt":".*","updatedAt":".*"\}\]\n`,
			status:   http.StatusOK,
		},
		{
			name:     "HTTP GET accounts.Details invalid address",
			method:   "GET",
			url:      "/invalid-address",
			expected: `not a valid address\n`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "HTTP GET accounts.Details unknown address",
			method:   "GET",
			url:      "/0f7025fa05b578e3",
			expected: `account not found\n`,
			status:   http.StatusNotFound,
		},
		{
			name:     "HTTP GET accounts.Details known address",
			method:   "GET",
			url:      "/<address>",
			expected: `\{"address":".*","createdAt":".*","updatedAt":".*"\}\n`,
			status:   http.StatusOK,
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			replacer := strings.NewReplacer(
				"<address>", tempAcc.Address,
			)

			url := replacer.Replace(string(step.url))

			req, err := http.NewRequest(step.method, url, nil)
			if err != nil {
				t.Fatalf("Did not expect an error, got: %s", err)
			}

			req.Context()

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			if status := rr.Code; status != step.status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, step.status)
			}

			// Store the new account if this test case created one
			if step.status == http.StatusCreated {
				json.Unmarshal(rr.Body.Bytes(), &tempAcc)
			}

			// Check the response body is what we expect.
			re := regexp.MustCompile(step.expected)
			match := re.FindString(rr.Body.String())
			if match == "" || match != rr.Body.String() {
				t.Errorf("handler returned unexpected body: got %q want %v", rr.Body.String(), re)
			}
		})
	}

	account.TeardDownTestService()
}
