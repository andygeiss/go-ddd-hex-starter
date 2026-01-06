package inbound_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_FileReader_ReadFileInfos_With_Current_Path_Should_Return_Two_Files(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	r := inbound.NewFileReader()
	wanted := 2
	for i := range wanted {
		_, _ = os.Create(fmt.Sprintf("%s/file%d.txt", tempDir, i))
	}
	ctx := context.Background()

	// Act
	fileInfos, err := r.ReadFileInfos(ctx, tempDir)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "fileInfos length must be 2", len(fileInfos), 2)
}

// Every benchmark should follow the Benchmark_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the benchmarks to be easy to read and understand.

// We use the following benchmarks to create a baseline for our application.
// This will be used for Profile Guided Optimization (PGO) later.
// You can run these benchmarks with `just profile` and writes the results to `.cpuprofile.out`.

// During the build process, we will use the results of these benchmarks to optimize our application.
// You can run the build process with `just build` and it will use the -pgo flag to optimize the application.

const (
	BenchmarkMaxFileCount = 1000
	BenchmarkMaxFileSize  = 1024 * 1024 // 1 MB
)

func Benchmark_FileReader_ReadFileInfos_With_1000_Entries_Should_Be_Fast(b *testing.B) {
	// Arrange
	tempDir := b.TempDir()
	reader := inbound.NewFileReader()
	ctx := context.Background()

	// Create BenchmarkMaxFileCount number of files in tempDir.
	// Each file should have a unique name and a random content.
	// Each file must be BenchmarkMaxFileSize bytes long.
	// We use a range loop over an int to create the files.
	for i := range BenchmarkMaxFileCount {
		file, err := os.Create(fmt.Sprintf("%s/file%d.txt", tempDir, i))
		if err != nil {
			b.Fatal(err)
		}
		defer func() { _ = file.Close() }()
		if _, err := file.Write(make([]byte, BenchmarkMaxFileSize)); err != nil {
			b.Fatal(err)
		}
	}

	// Reset the timer before the benchmark loop.
	b.ResetTimer()

	// We use b.Loop() to iterate over the benchmark loop.
	// This is the new way to iterate over the benchmark loop.
	for b.Loop() {
		// Read all generated file infos.
		fileInfos, err := reader.ReadFileInfos(ctx, tempDir)
		if err != nil {
			b.Fatal(err)
		}

		// Check if the number of file infos is correct.
		if len(fileInfos) != BenchmarkMaxFileCount {
			b.Fatalf("expected %d file infos, got %d", BenchmarkMaxFileCount, len(fileInfos))
		}
	}
}
