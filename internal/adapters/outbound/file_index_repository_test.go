package outbound_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// ============================================================================
// Functional Tests
// ============================================================================
// These tests verify the FileIndexRepository behaves correctly with real file I/O.

func Test_FileIndexRepository_Create_With_Valid_Index_Should_Persist_To_File(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_create.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/test/path")
	fileInfos := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
		{AbsPath: "/test/file2.txt", Size: 200},
	}
	index := indexing.NewIndex(id, fileInfos)

	// Act
	err := repo.Create(ctx, id, index)

	// Assert
	assert.That(t, "create error must be nil", err == nil, true)

	// Verify file was created
	_, statErr := os.Stat(path)
	assert.That(t, "file must exist", statErr == nil, true)
}

func Test_FileIndexRepository_Read_With_Existing_Index_Should_Return_Index(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_read.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/test/path")
	fileInfos := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
		{AbsPath: "/test/file2.txt", Size: 200},
		{AbsPath: "/test/file3.txt", Size: 300},
	}
	index := indexing.NewIndex(id, fileInfos)
	_ = repo.Create(ctx, id, index)

	// Act
	readIndex, err := repo.Read(ctx, id)

	// Assert
	assert.That(t, "read error must be nil", err == nil, true)
	assert.That(t, "file count must match", len(readIndex.FileInfos), 3)
	assert.That(t, "first file path must match", readIndex.FileInfos[0].AbsPath, "/test/file1.txt")
	assert.That(t, "first file size must match", readIndex.FileInfos[0].Size, int64(100))
}

func Test_FileIndexRepository_Read_With_NonExistent_Index_Should_Return_Error(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_nonexistent.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/nonexistent/path")

	// Act
	_, err := repo.Read(ctx, id)

	// Assert
	assert.That(t, "read error must not be nil", err != nil, true)
}

func Test_FileIndexRepository_Update_With_Existing_Index_Should_Modify_Data(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_update.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/test/path")
	originalFileInfos := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
	}
	originalIndex := indexing.NewIndex(id, originalFileInfos)
	_ = repo.Create(ctx, id, originalIndex)

	// Update with more files
	updatedFileInfos := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
		{AbsPath: "/test/file2.txt", Size: 200},
		{AbsPath: "/test/file3.txt", Size: 300},
	}
	updatedIndex := indexing.NewIndex(id, updatedFileInfos)

	// Act
	err := repo.Update(ctx, id, updatedIndex)

	// Assert
	assert.That(t, "update error must be nil", err == nil, true)

	// Verify update persisted
	readIndex, _ := repo.Read(ctx, id)
	assert.That(t, "file count must be updated", len(readIndex.FileInfos), 3)
}

func Test_FileIndexRepository_Delete_With_Existing_Index_Should_Remove_Data(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_delete.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/test/path")
	fileInfos := []indexing.FileInfo{
		{AbsPath: "/test/file1.txt", Size: 100},
	}
	index := indexing.NewIndex(id, fileInfos)
	_ = repo.Create(ctx, id, index)

	// Act
	err := repo.Delete(ctx, id)

	// Assert
	assert.That(t, "delete error must be nil", err == nil, true)

	// Verify index no longer readable
	_, readErr := repo.Read(ctx, id)
	assert.That(t, "read after delete must return error", readErr != nil, true)
}

func Test_FileIndexRepository_Create_With_Empty_FileInfos_Should_Persist(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_empty.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/empty/path")
	index := indexing.NewIndex(id, []indexing.FileInfo{})

	// Act
	err := repo.Create(ctx, id, index)

	// Assert
	assert.That(t, "create error must be nil", err == nil, true)

	readIndex, readErr := repo.Read(ctx, id)
	assert.That(t, "read error must be nil", readErr == nil, true)
	assert.That(t, "file count must be zero", len(readIndex.FileInfos), 0)
}

func Test_FileIndexRepository_Create_With_Large_FileSize_Should_Handle_Correctly(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_large_size.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/large/path")
	largeSize := int64(1024 * 1024 * 1024 * 10) // 10 GB
	fileInfos := []indexing.FileInfo{
		{AbsPath: "/large/file.bin", Size: largeSize},
	}
	index := indexing.NewIndex(id, fileInfos)

	// Act
	err := repo.Create(ctx, id, index)

	// Assert
	assert.That(t, "create error must be nil", err == nil, true)

	readIndex, _ := repo.Read(ctx, id)
	assert.That(t, "large file size must be preserved", readIndex.FileInfos[0].Size, largeSize)
}

func Test_FileIndexRepository_Hash_Should_Be_Consistent_After_Roundtrip(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_hash.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/hash/path")
	fileInfos := []indexing.FileInfo{
		{AbsPath: "/hash/file1.txt", Size: 100},
		{AbsPath: "/hash/file2.txt", Size: 200},
	}
	index := indexing.NewIndex(id, fileInfos)
	originalHash := index.Hash()

	// Act
	_ = repo.Create(ctx, id, index)
	readIndex, _ := repo.Read(ctx, id)
	readHash := readIndex.Hash()

	// Assert
	assert.That(t, "hash must be consistent after roundtrip", readHash, originalHash)
}

func Test_FileIndexRepository_Create_With_Existing_ID_Should_Return_Error(t *testing.T) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()

	path := "testdata/test_duplicate.json"
	repo := outbound.NewFileIndexRepository(path)
	ctx := context.Background()

	id := indexing.IndexID("/duplicate/path")

	// First create
	firstFileInfos := []indexing.FileInfo{
		{AbsPath: "/first/file.txt", Size: 100},
	}
	firstIndex := indexing.NewIndex(id, firstFileInfos)
	_ = repo.Create(ctx, id, firstIndex)

	// Second create with same ID
	secondFileInfos := []indexing.FileInfo{
		{AbsPath: "/second/file.txt", Size: 999},
	}
	secondIndex := indexing.NewIndex(id, secondFileInfos)

	// Act
	err := repo.Create(ctx, id, secondIndex)

	// Assert - Create should fail for existing ID (use Update instead)
	assert.That(t, "create with existing ID must return error", err != nil, true)

	// Original data should be unchanged
	readIndex, _ := repo.Read(ctx, id)
	assert.That(t, "file path must be from first create", readIndex.FileInfos[0].AbsPath, "/first/file.txt")
}

// ============================================================================
// Benchmarks
// ============================================================================
// These benchmarks create a baseline for Profile-Guided Optimization (PGO).
// Run with `just profile` to generate cpuprofile.pprof.

const (
	BenchmarkMaxFileCount = 1000
	BenchmarkMaxFileSize  = 1024 * 1024 // 1 MB
)

func Benchmark_FileIndexRepository_Create_With_1000_Entries_Should_Be_Fast(b *testing.B) {
	// Arrange
	_ = os.MkdirAll("testdata", 0755)
	defer func() { _ = os.RemoveAll("testdata") }()
	path := "testdata/index.json"
	repo := outbound.NewFileIndexRepository(path)

	// Create BenchmarkMaxFileCount number of indexing.FileInfo as a slice.
	// This will be used to create the index and benchmark the performance.
	files := make([]indexing.FileInfo, BenchmarkMaxFileCount)

	// Use range loop over int to initialize the slice
	for i := range files {
		files[i] = indexing.FileInfo{
			AbsPath: fmt.Sprintf("file%d.txt", i),
			Size:    BenchmarkMaxFileSize,
		}
	}

	// Create the Index instance by using the path as the ID.
	id := indexing.IndexID(path)
	index := indexing.NewIndex(id, files)
	ctx := context.Background()

	// Benchmark
	b.ResetTimer()
	for b.Loop() {
		_ = repo.Create(ctx, id, index)
	}
}
