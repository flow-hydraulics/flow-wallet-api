package tests

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/flow-hydraulics/flow-wallet-api/handlers"
	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
	"github.com/gorilla/mux"
)

func TestSettingsE2E(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svcs := test.GetServices(t, cfg)

	sysHandler := handlers.NewSystem(svcs.GetSystem())

	router := mux.NewRouter()
	router.Handle("/settings", sysHandler.GetSettings()).Methods(http.MethodGet)
	router.Handle("/settings", sysHandler.SetSettings()).Methods(http.MethodPost)

	var steps = []struct {
		body           io.Reader
		path           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			body:           nil,
			path:           "/settings",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"maintenanceMode":false}`,
		},
		{
			body:           bytes.NewBufferString("{\"maintenanceMode\": true}"),
			path:           "/settings",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"maintenanceMode":true}`,
		},
		{
			body:           nil,
			path:           "/settings",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"maintenanceMode":true}`,
		},
	}

	for _, tt := range steps {
		res := send(router, tt.method, tt.path, tt.body)
		assertStatusCode(t, res, tt.expectedStatus)
		if bs, err := ioutil.ReadAll(res.Body); err != nil || strings.TrimSpace(string(bs)) != tt.expectedBody {
			if err != nil {
				t.Error(err)
			} else {
				t.Errorf("expected response body to equal '%v', got '%v'\n", tt.expectedBody, strings.TrimSpace(string(bs)))
			}
		}
	}
}

func TestIsMaintenanceMode(t *testing.T) {
	cfg := test.LoadConfig(t, testConfigPath)
	svcs := test.GetServices(t, cfg)

	sysService := svcs.GetSystem()

	if sysService.IsMaintenanceMode() {
		t.Error("expected system not to be in maintenance mode")
	}

	settings, err := sysService.GetSettings()
	if err != nil {
		t.Fatal(err)
	}

	settings.MaintenanceMode = true

	if err := sysService.SaveSettings(settings); err != nil {
		t.Fatal(err)
	}

	if !sysService.IsMaintenanceMode() {
		t.Error("expected system to be in maintenance mode")
	}
}
