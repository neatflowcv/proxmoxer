package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/neatflowcv/proxmoxer/internal/application/dto"
	"github.com/neatflowcv/proxmoxer/internal/domain/common"
)

// ResponseWriter wraps common response writing functionality.
type ResponseWriter struct {
	logger *log.Logger
}

// NewResponseWriter creates a new ResponseWriter.
func NewResponseWriter(logger *log.Logger) *ResponseWriter {
	return &ResponseWriter{logger: logger}
}

// WriteJSON writes a JSON response.
func (rw *ResponseWriter) WriteJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

// WriteError writes an error response.
func (rw *ResponseWriter) WriteError(w http.ResponseWriter, statusCode int, message string, details ...string) error {
	var detailsMap map[string]any
	if len(details) > 0 {
		detailsMap = map[string]any{
			"details": details,
		}
	}

	errResp := dto.ErrorResponse{
		Code:    http.StatusText(statusCode),
		Message: message,
		Details: detailsMap,
	}

	return rw.WriteJSON(w, statusCode, errResp)
}

// HandleError handles different types of errors and writes appropriate responses.
func (rw *ResponseWriter) HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	// Determine status code and message based on error type
	statusCode := http.StatusInternalServerError
	message := "An internal error occurred"

	switch {
	case errors.Is(err, common.ErrClusterNotFound):
		statusCode = http.StatusNotFound
		message = "Cluster not found"
	case errors.Is(err, common.ErrClusterAlreadyExists):
		statusCode = http.StatusConflict
		message = "Cluster already exists"
	case errors.Is(err, common.ErrInvalidClusterID):
		statusCode = http.StatusBadRequest
		message = "Invalid cluster ID"
	case errors.Is(err, common.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		message = "Invalid credentials"
	case errors.Is(err, common.ErrAuthenticationFailed):
		statusCode = http.StatusUnauthorized
		message = "Authentication failed"
	case errors.Is(err, common.ErrProxmoxConnectionFailed):
		statusCode = http.StatusBadGateway
		message = "Failed to connect to Proxmox"
	}

	rw.logger.Printf("[ERROR] Handling error: %v (status: %d)\n", err, statusCode)

	writeErr := rw.WriteError(w, statusCode, message)
	if writeErr != nil {
		rw.logger.Printf("[ERROR] Failed to write error response: %v\n", writeErr)
	}
}
