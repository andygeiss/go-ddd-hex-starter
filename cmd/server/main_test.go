package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"

	"github.com/andygeiss/cloud-native-utils/logging"
)

// Every benchmark should follow the Benchmark_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the benchmarks to be easy to read and understand.

// We use the following benchmarks to create a baseline for our application.
// This will be used for Profile Guided Optimization (PGO) later.
// You can run these benchmarks with `just profile` and writes the results to `.cpuprofile.out`.

// During the build process, we will use the results of these benchmarks to optimize our application.
// You can run the build process with `just build` and it will use the -pgo flag to optimize the application.

// Benchmark_Server_Route_With_Liveness_Endpoint_Should_Respond_Efficiently benchmarks the /liveness endpoint.
// This endpoint is critical for Kubernetes health checks and must be fast.
func Benchmark_Server_Route_With_Liveness_Endpoint_Should_Respond_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template.
	req := httptest.NewRequest(http.MethodGet, "/liveness", nil)

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

// Benchmark_Server_Route_With_Readiness_Endpoint_Should_Respond_Efficiently benchmarks the /readiness endpoint.
// This endpoint is critical for Kubernetes readiness checks and must be fast.
func Benchmark_Server_Route_With_Readiness_Endpoint_Should_Respond_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template.
	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

// Benchmark_Server_Route_With_Login_View_Should_Render_Efficiently benchmarks the /ui/login endpoint.
// This endpoint renders the login template and is critical for first impressions.
func Benchmark_Server_Route_With_Login_View_Should_Render_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template.
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

// Benchmark_Server_Route_With_Static_CSS_Should_Serve_Efficiently benchmarks serving static CSS assets.
// Static assets like CSS are served frequently and must be optimized.
func Benchmark_Server_Route_With_Static_CSS_Should_Serve_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template for a CSS file.
	req := httptest.NewRequest(http.MethodGet, "/static/css/styles.css", nil)

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

// Benchmark_Server_Route_With_Static_JS_Should_Serve_Efficiently benchmarks serving static JS assets.
// Static assets like JavaScript are served frequently and must be optimized.
func Benchmark_Server_Route_With_Static_JS_Should_Serve_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template for a JS file.
	req := httptest.NewRequest(http.MethodGet, "/static/js/htmx.min.js", nil)

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

// Benchmark_Server_Route_With_UI_Index_Unauthenticated_Should_Redirect_Efficiently benchmarks
// the /ui/ endpoint for unauthenticated requests. This tests the redirect path performance.
func Benchmark_Server_Route_With_UI_Index_Unauthenticated_Should_Redirect_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template for the index view (unauthenticated).
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}

// Benchmark_Server_Route_With_Mixed_Workload_Should_Handle_Efficiently benchmarks a realistic
// mixed workload of different endpoint types to simulate production traffic patterns.
func Benchmark_Server_Route_With_Mixed_Workload_Should_Handle_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Define the endpoints to hit in a mixed workload.
	endpoints := []string{
		"/liveness",
		"/readiness",
		"/ui/login",
		"/static/css/styles.css",
	}

	// Create requests for all endpoints.
	requests := make([]*http.Request, len(endpoints))
	for i, endpoint := range endpoints {
		requests[i] = httptest.NewRequest(http.MethodGet, endpoint, nil)
	}

	b.ResetTimer()
	for b.Loop() {
		// Hit each endpoint in round-robin fashion.
		for _, req := range requests {
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
		}
	}
}

// Benchmark_Server_Route_With_Templating_Engine_Should_Render_Efficiently benchmarks
// the templating engine by rendering the login template multiple times.
// This benchmark focuses on the template rendering path.
func Benchmark_Server_Route_With_Templating_Engine_Should_Render_Efficiently(b *testing.B) {
	// Setup the server context and mux once before the benchmark.
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)

	// Create the request template for the login view.
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)

	b.ResetTimer()
	for b.Loop() {
		// Use a new recorder for each iteration to avoid state interference.
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
	}
}
