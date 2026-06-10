package api

import (
	"log"
	"net/http"
	"strings"
)

// adminRole is the Entra ID app-role value that grants administrative access
// (defined on the lp-production-planning-issues-workbench app registration).
const adminRole = "admin"

// adminMiddleware checks if the signed-in user carries the admin app role.
// Roles come from the Entra ID token's `roles` claim, captured in the
// session at login.
func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.sessionStore.Get(r, "m3-session")

		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		rawRoles, _ := session.Values["user_roles"].(string)
		hasAdminRole := false
		for _, role := range strings.Split(rawRoles, ",") {
			if strings.TrimSpace(role) == adminRole {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			log.Printf("WARN: User %v attempted to access admin endpoint without the %q app role", session.Values["user_email"], adminRole)
			http.Error(w, "Forbidden: administrator role required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
