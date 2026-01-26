package api

import (
	"fmt"
	"log"
	"net/http"
)

// adminMiddleware checks if the user has system administrator role
func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: adminMiddleware called for path: %s", r.URL.Path)

		// Get user ID from session
		userID, err := s.getUserIDFromSession(r)
		if err != nil {
			log.Printf("ERROR: Failed to get user ID in adminMiddleware: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("DEBUG: Checking admin role for user: %s", userID)

		// Check if user has admin role (Infor-SystemAdministrator)
		hasAdminRole, err := s.userProfileService.HasRole(r.Context(), userID, "Infor-SystemAdministrator")
		if err != nil {
			log.Printf("ERROR: Failed to check admin role: %v", err)
			http.Error(w, fmt.Sprintf("Failed to check permissions: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("DEBUG: User %s has admin role: %v", userID, hasAdminRole)

		if !hasAdminRole {
			log.Printf("WARN: User %s attempted to access admin endpoint without permission", userID)
			http.Error(w, "Forbidden: System administrator role required", http.StatusForbidden)
			return
		}

		log.Printf("DEBUG: Admin check passed for user: %s", userID)
		next.ServeHTTP(w, r)
	})
}
