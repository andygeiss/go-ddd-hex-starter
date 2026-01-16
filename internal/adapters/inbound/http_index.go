package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/redirecting"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewIndexResponse specifies the view data.
type HttpViewIndexResponse struct {
	AppName   string
	Email     string
	Issuer    string
	Name      string
	SessionID string
	Subject   string
	Title     string
	Verified  bool
}

// HttpViewIndex defines an HTTP handler function for rendering the index template.
func HttpViewIndex(e *templating.Engine) http.HandlerFunc {
	// Retrieve application details from environment variables at startup.
	// We can reuse these values instead of reading them from the environment on each request.
	appName := os.Getenv("APP_NAME")
	title := appName + " - " + os.Getenv("APP_DESCRIPTION")

	return func(w http.ResponseWriter, r *http.Request) {
		// Make a shortcut for the current context.
		ctx := r.Context()

		// Check if the user is authenticated.
		// We check both sessionID and email because:
		// - sessionID might exist (from cookie) even after logout
		// - email being empty indicates the session was deleted server-side
		sessionID, _ := ctx.Value(security.ContextSessionID).(string)
		email, _ := ctx.Value(security.ContextEmail).(string)
		if sessionID == "" || email == "" {
			redirecting.Redirect(w, r, "/ui/login")
			return
		}

		// Add session-specific data.
		data := HttpViewIndexResponse{
			AppName:   appName,
			Email:     email,
			Issuer:    ctx.Value(security.ContextIssuer).(string),
			Name:      ctx.Value(security.ContextName).(string),
			SessionID: sessionID,
			Subject:   ctx.Value(security.ContextSubject).(string),
			Title:     title,
			Verified:  ctx.Value(security.ContextVerified).(bool),
		}

		// Render the template using the provided engine and data.
		HttpView(e, "index", data)(w, r)
	}
}
