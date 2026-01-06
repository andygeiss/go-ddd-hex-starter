package main

import (
	"context"
	"os"
	"testing"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// Benchmark for Profile-Guided Optimization (PGO).
// Run with: just profile
// This generates cpuprofile.pprof for optimized builds.

func Benchmark_CLI_Integration_With_Kafka_Should_Index_Efficiently(b *testing.B) {
	// Skip if KAFKA_BROKERS is not set.
	if os.Getenv("KAFKA_BROKERS") == "" {
		b.Skip("Skipping benchmark: KAFKA_BROKERS not set")
	}

	ctx := context.Background()
	dispatcher := messaging.NewExternalDispatcher()
	fileReader := inbound.NewFileReader()

	indexPath := "./bench_index.json"
	defer func() { _ = os.Remove(indexPath) }()
	indexRepository := outbound.NewFileIndexRepository(indexPath)
	eventPublisher := outbound.NewEventPublisher(dispatcher)

	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	wd, _ := os.Getwd()

	for b.Loop() {
		_ = indexingService.CreateIndex(ctx, wd)
	}
}
