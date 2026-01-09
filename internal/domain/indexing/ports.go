package indexing

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/resource"
)

// FileReader represents the interface for interacting with the filesystem.
// It is responsible for reading file information from the filesystem.
// We use context.Context to provide cancellation and timeout capabilities.
type FileReader interface {
	ReadFileInfos(ctx context.Context, path string) ([]FileInfo, error)
}

// IndexRepository represents the repository for indexing.
// It provides methods for creating, retrieving, updating, deleting, and listing indexes.
// We will not reinvent the wheel and use the resource.Access type from the cloud-native-utils package.
type IndexRepository resource.Access[IndexID, Index]
