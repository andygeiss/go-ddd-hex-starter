package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// Benchmarks for Profile-Guided Optimization (PGO).
// Run with: just profile
// This generates cpuprofile.pprof for optimized builds.

func Benchmark_Server_Integration_Liveness_Should_Respond_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/liveness")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_Static_CSS_Should_Serve_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/static/css/base.css")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}

func Benchmark_Server_Integration_Login_Page_Should_Render_Fast(b *testing.B) {
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, efs, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for b.Loop() {
		resp, _ := client.Get(server.URL + "/ui/login")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
}
