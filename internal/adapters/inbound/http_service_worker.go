package inbound

import (
	"net/http"
	"os"

	"github.com/andygeiss/cloud-native-utils/templating"
)

// HttpViewServiceWorkerResponse specifies the view data for the service worker.
type HttpViewServiceWorkerResponse struct {
	CacheName string
	Version   string
}

// HttpViewServiceWorker defines an HTTP handler function for serving the service worker.
// The service worker enables PWA features like offline caching and installability.
func HttpViewServiceWorker(e *templating.Engine) http.HandlerFunc {
	// Retrieve application details from environment variables at startup.
	// We can reuse these values instead of reading them from the environment on each request.
	appName := os.Getenv("APP_NAME")
	appVersion := os.Getenv("APP_VERSION")

	// Create the Data Object (DTO) once at startup.
	data := HttpViewServiceWorkerResponse{
		CacheName: appName,
		Version:   appVersion,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JavaScript for the service worker.
		w.Header().Set("Content-Type", "application/javascript")
		// Service workers must not be cached to ensure updates are applied.
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		HttpView(e, "sw", data)(w, r)
	}
}
