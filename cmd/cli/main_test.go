package main

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// ============================================================================
// Unit Tests (no external dependencies)
// ============================================================================
// These tests use mocks and can run without any external services.

func Test_CLI_Unit_IndexingService_CreateIndex_With_Mock_Should_Succeed(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockFileReader := &mockFileReader{
		fileInfos: []indexing.FileInfo{
			{AbsPath: "/test/file1.txt", Size: 100},
			{AbsPath: "/test/file2.txt", Size: 200},
		},
	}
	mockIndexRepository := resource.NewMockAccess[indexing.IndexID, indexing.Index]()
	mockIndexRepository.WithCreateFn(func(ctx context.Context, id indexing.IndexID, index indexing.Index) error {
		return nil
	})
	mockEventPublisher := &mockEventPublisher{}

	indexingService := indexing.NewIndexingService(mockFileReader, mockIndexRepository, mockEventPublisher)

	// Act
	err := indexingService.CreateIndex(ctx, "/test/path")

	// Assert
	assert.That(t, "create index error must be nil", err == nil, true)
	assert.That(t, "event must be published", mockEventPublisher.Published, true)
}

func Test_CLI_Unit_IndexingService_IndexFiles_With_Mock_Should_Return_Files(t *testing.T) {
	// Arrange
	ctx := context.Background()
	expectedFiles := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
		{AbsPath: "/test/file2.txt", Size: 200},
		{AbsPath: "/test/file3.txt", Size: 300},
	}
	mockFileReader := &mockFileReader{fileInfos: expectedFiles}
	mockIndexRepository := resource.NewMockAccess[indexing.IndexID, indexing.Index]()
	mockIndexRepository.WithCreateFn(func(ctx context.Context, id indexing.IndexID, index indexing.Index) error {
		return nil
	})
	mockEventPublisher := &mockEventPublisher{}

	indexingService := indexing.NewIndexingService(mockFileReader, mockIndexRepository, mockEventPublisher)

	// First create the index
	_ = indexingService.CreateIndex(ctx, "/test/path")

	// Act
	files, err := indexingService.IndexFiles(ctx, "/test/path")

	// Assert
	assert.That(t, "index files error must be nil", err == nil, true)
	assert.That(t, "files count must match", len(files), len(expectedFiles))
}

func Test_CLI_Unit_FileReader_ReadFileInfos_With_TestData_Should_Return_Files(t *testing.T) {
	// Arrange
	ctx := context.Background()
	_ = os.MkdirAll("testdata_cli", 0755)
	defer func() { _ = os.RemoveAll("testdata_cli") }()

	// Create test files
	for i := range 3 {
		f, _ := os.Create("testdata_cli/file" + string(rune('a'+i)) + ".txt")
		_, _ = f.WriteString("test content")
		_ = f.Close()
	}

	fileReader := inbound.NewFileReader()

	// Act
	fileInfos, err := fileReader.ReadFileInfos(ctx, "testdata_cli")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 3 files", len(fileInfos), 3)
}

func Test_CLI_Unit_FileIndexRepository_Create_And_Read_Should_Work(t *testing.T) {
	// Arrange
	ctx := context.Background()
	indexPath := "./test_cli_index.json"
	defer func() { _ = os.Remove(indexPath) }()

	repository := outbound.NewFileIndexRepository(indexPath)
	id := indexing.IndexID("/test/path")
	fileInfos := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
	}
	index := indexing.NewIndex(id, fileInfos)

	// Act
	createErr := repository.Create(ctx, id, index)
	readIndex, readErr := repository.Read(ctx, id)

	// Assert
	assert.That(t, "create error must be nil", createErr == nil, true)
	assert.That(t, "read error must be nil", readErr == nil, true)
	assert.That(t, "file count must match", len(readIndex.FileInfos), 1)
}

// Mock implementations for unit tests.
type mockFileReader struct {
	fileInfos []indexing.FileInfo
}

func (m *mockFileReader) ReadFileInfos(ctx context.Context, path string) ([]indexing.FileInfo, error) {
	return m.fileInfos, nil
}

type mockEventPublisher struct {
	Published bool
}

func (m *mockEventPublisher) Publish(ctx context.Context, e event.Event) error {
	m.Published = true
	return nil
}

// ============================================================================
// Integration tests for the CLI application with external Kafka dispatcher.
// ============================================================================
// These tests require a running Kafka instance (KAFKA_BROKERS environment variable).
//
// Run with: KAFKA_BROKERS=localhost:9092 go test -v ./cmd/cli/...
// Or via Docker Compose: just up && just test

func Test_CLI_Integration_With_Kafka_Should_Publish_And_Receive_Event(t *testing.T) {
	// Skip if KAFKA_BROKERS is not set (Kafka not available).
	if os.Getenv("KAFKA_BROKERS") == "" {
		t.Skip("Skipping integration test: KAFKA_BROKERS not set")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dispatcher := messaging.NewExternalDispatcher()
	fileReader := inbound.NewFileReader()
	eventSubscriber := inbound.NewEventSubscriber(dispatcher)

	indexPath := "./test_index.json"
	defer func() { _ = os.Remove(indexPath) }()
	indexRepository := outbound.NewFileIndexRepository(indexPath)
	eventPublisher := outbound.NewEventPublisher(dispatcher)

	// Track event reception.
	var eventReceived atomic.Bool
	var receivedFileCount int

	err := eventSubscriber.Subscribe(
		ctx,
		indexing.EventTopicFileIndexCreated,
		func() event.Event { return indexing.NewEventFileIndexCreated() },
		func(e event.Event) error {
			evt := e.(*indexing.EventFileIndexCreated)
			receivedFileCount = evt.FileCount
			eventReceived.Store(true)
			return nil
		},
	)
	assert.That(t, "subscribe error must be nil", err == nil, true)

	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	// Act
	wd, err := os.Getwd()
	assert.That(t, "getwd error must be nil", err == nil, true)

	err = indexingService.CreateIndex(ctx, wd)
	assert.That(t, "create index error must be nil", err == nil, true)

	// Wait for async event processing.
	time.Sleep(200 * time.Millisecond)

	// Assert
	assert.That(t, "event must be received", eventReceived.Load(), true)
	assert.That(t, "file count must be greater than 0", receivedFileCount > 0, true)

	// Verify index was persisted.
	id := indexing.IndexID(wd)
	index, err := indexRepository.Read(ctx, id)
	assert.That(t, "read index error must be nil", err == nil, true)
	assert.That(t, "index file count must match event", len(index.FileInfos), receivedFileCount)
}

func Test_CLI_Integration_With_Kafka_Should_Create_Valid_Index_Hash(t *testing.T) {
	// Skip if KAFKA_BROKERS is not set (Kafka not available).
	if os.Getenv("KAFKA_BROKERS") == "" {
		t.Skip("Skipping integration test: KAFKA_BROKERS not set")
	}

	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dispatcher := messaging.NewExternalDispatcher()
	fileReader := inbound.NewFileReader()

	indexPath := "./test_index_hash.json"
	defer func() { _ = os.Remove(indexPath) }()
	indexRepository := outbound.NewFileIndexRepository(indexPath)
	eventPublisher := outbound.NewEventPublisher(dispatcher)

	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	// Act
	wd, err := os.Getwd()
	assert.That(t, "getwd error must be nil", err == nil, true)

	err = indexingService.CreateIndex(ctx, wd)
	assert.That(t, "create index error must be nil", err == nil, true)

	// Assert
	id := indexing.IndexID(wd)
	index, err := indexRepository.Read(ctx, id)
	assert.That(t, "read index error must be nil", err == nil, true)

	hash := index.Hash()
	assert.That(t, "hash must not be empty", len(hash) > 0, true)
	assert.That(t, "hash must be 64 characters (SHA256 hex)", len(hash), 64)
}

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
