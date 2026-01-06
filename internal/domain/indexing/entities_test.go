package indexing_test

import (
	"testing"
	"time"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_FileInfo_NewFileInfo_With_Valid_Data_Should_Create_Instance(t *testing.T) {
	// Arrange
	absPath := "/path/to/file.txt"
	size := int64(1024)
	modTime := time.Now()

	// Act
	fileInfo := indexing.NewFileInfo(absPath, size, modTime)

	// Assert
	assert.That(t, "AbsPath must match", fileInfo.AbsPath, absPath)
	assert.That(t, "Size must match", fileInfo.Size, size)
	assert.That(t, "ModTime must match", fileInfo.ModTime, modTime)
}

func Test_FileInfo_NewFileInfo_With_Empty_Path_Should_Create_Instance(t *testing.T) {
	// Arrange
	absPath := ""
	size := int64(0)
	modTime := time.Time{}

	// Act
	fileInfo := indexing.NewFileInfo(absPath, size, modTime)

	// Assert
	assert.That(t, "AbsPath must be empty", fileInfo.AbsPath, absPath)
	assert.That(t, "Size must be zero", fileInfo.Size, size)
	assert.That(t, "ModTime must be zero value", fileInfo.ModTime, modTime)
}

func Test_FileInfo_NewFileInfo_With_Large_Size_Should_Handle_Correctly(t *testing.T) {
	// Arrange
	absPath := "/very/long/path/to/a/deeply/nested/file.txt"
	size := int64(1024 * 1024 * 1024 * 10) // 10 GB
	modTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	// Act
	fileInfo := indexing.NewFileInfo(absPath, size, modTime)

	// Assert
	assert.That(t, "AbsPath must match", fileInfo.AbsPath, absPath)
	assert.That(t, "Size must handle large values", fileInfo.Size, size)
	assert.That(t, "ModTime must match", fileInfo.ModTime, modTime)
}
