package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewManifestResponse specifies the view data for the PWA manifest.
type HttpViewManifestResponse struct {
	Description string
	Name        string
	ShortName   string
}

// HttpViewManifest defines an HTTP handler function for rendering the PWA manifest.json.
func HttpViewManifest(e *templating.Engine) http.HandlerFunc {
	// Retrieve application details from environment variables at startup.
	// We can reuse these values instead of reading them from the environment on each request.
	appName := os.Getenv("APP_NAME")
	description := os.Getenv("APP_DESCRIPTION")

	// Create the Data Object (DTO) once at startup.
	data := HttpViewManifestResponse{
		Description: description,
		Name:        appName,
		ShortName:   appName,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to application/manifest+json for PWA manifest.
		w.Header().Set("Content-Type", "application/manifest+json")
		HttpView(e, "manifest", data)(w, r)
	}
}
