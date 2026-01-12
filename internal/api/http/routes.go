package http

import (
	"log"
	"net/http"

	"github.com/neatflowcv/proxmoxer/internal/api/http/handler"
	"github.com/neatflowcv/proxmoxer/internal/api/http/middleware"
	"github.com/neatflowcv/proxmoxer/internal/application/services"
)

// Router sets up HTTP routes for the API.
type Router struct {
	mux            *http.ServeMux
	clusterHandler *handler.ClusterHandler
	logger         *log.Logger
}

// NewRouter creates a new Router with all handlers.
func NewRouter(
	clusterService *services.ClusterService,
	logger *log.Logger,
) *Router {
	if logger == nil {
		logger = log.Default()
	}

	router := &Router{
		mux:            http.NewServeMux(),
		clusterHandler: handler.NewClusterHandler(clusterService, logger),
		logger:         logger,
	}

	router.setupRoutes()

	return router
}

// Mux returns the underlying HTTP multiplexer.
func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

// ServeHTTP makes Router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	middleware.CORS(r.mux).ServeHTTP(w, req)
}

// setupRoutes registers all API routes.
func (r *Router) setupRoutes() {
	r.logger.Println("Setting up API routes")

	// Cluster routes
	// POST /api/v1/clusters - Register a new cluster
	r.mux.HandleFunc("POST /api/v1/clusters", r.clusterHandler.RegisterCluster)

	// GET /api/v1/clusters - List all clusters
	r.mux.HandleFunc("GET /api/v1/clusters", r.clusterHandler.ListClusters)

	// GET /api/v1/clusters/{id} - Get a specific cluster
	r.mux.HandleFunc("GET /api/v1/clusters/{id}", r.clusterHandler.GetCluster)

	// DELETE /api/v1/clusters/{id} - Deregister a cluster
	r.mux.HandleFunc("DELETE /api/v1/clusters/{id}", r.clusterHandler.DeregisterCluster)

	// GET /api/v1/clusters/{id}/disks - Get disk information for all nodes in a cluster
	r.mux.HandleFunc("GET /api/v1/clusters/{id}/disks", r.clusterHandler.ListClusterDisks)

	// Health check endpoint
	logger := r.logger
	r.mux.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{"status":"healthy"}`))
		if err != nil {
			logger.Printf("Failed to write health check response: %v\n", err)
		}
	})

	r.logger.Println("API routes configured successfully")
}
