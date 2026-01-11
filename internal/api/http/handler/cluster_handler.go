package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/neatflowcv/proxmoxer/internal/application/dto"
	"github.com/neatflowcv/proxmoxer/internal/application/services"
)

// ClusterHandler handles HTTP requests for cluster operations
type ClusterHandler struct {
	clusterService *services.ClusterService
	responseWriter *ResponseWriter
	logger         *log.Logger
}

// NewClusterHandler creates a new ClusterHandler
func NewClusterHandler(
	clusterService *services.ClusterService,
	logger *log.Logger,
) *ClusterHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &ClusterHandler{
		clusterService: clusterService,
		responseWriter: NewResponseWriter(logger),
		logger:         logger,
	}
}

// RegisterCluster handles POST /api/v1/clusters
// Registers a new Proxmox cluster
func (h *ClusterHandler) RegisterCluster(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("[Handler] Handling RegisterCluster request")

	if r.Method != http.MethodPost {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req dto.RegisterClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			h.responseWriter.WriteError(w, http.StatusBadRequest, "Request body is required")
		} else {
			h.responseWriter.WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		}
		return
	}

	// Call service
	response, err := h.clusterService.RegisterCluster(r.Context(), &req)
	if err != nil {
		h.logger.Printf("[Handler] RegisterCluster service error: %v\n", err)
		h.responseWriter.HandleError(w, err)
		return
	}

	// Write success response
	h.responseWriter.WriteJSON(w, http.StatusCreated, response)
}

// ListClusters handles GET /api/v1/clusters
// Lists all registered clusters
func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("[Handler] Handling ListClusters request")

	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Call service
	response, err := h.clusterService.ListClusters(r.Context())
	if err != nil {
		h.logger.Printf("[Handler] ListClusters service error: %v\n", err)
		h.responseWriter.HandleError(w, err)
		return
	}

	// Write success response
	h.responseWriter.WriteJSON(w, http.StatusOK, response)
}

// GetCluster handles GET /api/v1/clusters/{id}
// Gets a specific cluster by ID
func (h *ClusterHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("[Handler] Handling GetCluster request")

	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract cluster ID from URL path
	// URL pattern: /api/v1/clusters/{id}
	clusterID := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")

	if clusterID == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	// Call service
	response, err := h.clusterService.GetCluster(r.Context(), clusterID)
	if err != nil {
		h.logger.Printf("[Handler] GetCluster service error: %v\n", err)
		h.responseWriter.HandleError(w, err)
		return
	}

	// Write success response
	h.responseWriter.WriteJSON(w, http.StatusOK, response)
}

// DeregisterCluster handles DELETE /api/v1/clusters/{id}
// Deregisters (removes) a cluster
func (h *ClusterHandler) DeregisterCluster(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("[Handler] Handling DeregisterCluster request")

	if r.Method != http.MethodDelete {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract cluster ID from URL path
	clusterID := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")

	if clusterID == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Cluster ID is required")
		return
	}

	// Call service
	err := h.clusterService.DeregisterCluster(r.Context(), clusterID)
	if err != nil {
		h.logger.Printf("[Handler] DeregisterCluster service error: %v\n", err)
		h.responseWriter.HandleError(w, err)
		return
	}

	// Write success response (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}
