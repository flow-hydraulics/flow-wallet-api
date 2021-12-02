package handlers

import (
	"encoding/json"

	"net/http"

	"github.com/flow-hydraulics/flow-wallet-api/system"
)

// System is a HTTP server for system settings management.
type System struct {
	service *system.Service
}

func NewSystem(service *system.Service) *System {
	return &System{service}
}

func (s *System) GetSettings() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		res, err := s.service.GetSettings()

		if err != nil {
			handleError(rw, r, err)
			return
		}

		handleJsonResponse(rw, http.StatusOK, res.ToJSON())
	})
}

func (s *System) SetSettings() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Check body is not empty
		if err := checkNonEmptyBody(r); err != nil {
			handleError(rw, r, err)
			return
		}

		// Get existing settings
		settings, err := s.service.GetSettings()
		if err != nil {
			handleError(rw, r, err)
			return
		}

		// Convert existing to JSON model
		settingsJSON := settings.ToJSON()

		// Decode JSON over existing settings
		// Should not change fields which do not exist in request body
		if err := json.NewDecoder(r.Body).Decode(&settingsJSON); err != nil {
			handleError(rw, r, InvalidBodyError)
			return
		}

		// TODO: Check if maintenance mode was enabled and logger it

		// Assign fields from JSON back to application model
		settings.FromJSON(settingsJSON)

		// Save in database
		if err := s.service.SaveSettings(settings); err != nil {
			handleError(rw, r, err)
			return
		}

		// Return updated version
		handleJsonResponse(rw, http.StatusOK, settings.ToJSON())
	})
}
