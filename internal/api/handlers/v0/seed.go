package v0

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/registry/internal/service"
)

// SeedHandler handles GET requests for the seed.json endpoint
// Returns all servers in the seed format for composability
func SeedHandler(registry service.RegistryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get all servers from the registry using List with empty cursor and high limit
		servers, _, err := registry.List("", 10000) // Use high limit to get all servers
		if err != nil {
			log.Printf("Error getting servers for seed export: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Convert servers to ServerDetail format for seed export
		// The seed format expects ServerDetail objects with full package/remote information
		serverDetails := make([]interface{}, len(servers))
		for i, server := range servers {
			// Get detailed information for each server
			detail, err := registry.GetByID(server.ID)
			if err != nil {
				log.Printf("Error getting server detail for %s: %v", server.ID, err)
				// Fall back to basic server information
				serverDetails[i] = server
				continue
			}
			serverDetails[i] = detail
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Encode and return the seed data
		if err := json.NewEncoder(w).Encode(serverDetails); err != nil {
			log.Printf("Error encoding seed data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
