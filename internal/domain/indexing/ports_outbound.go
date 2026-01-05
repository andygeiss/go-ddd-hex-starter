package indexing

import (
	"github.com/andygeiss/cloud-native-utils/resource"
)

// IndexRepository represents the repository for indexing.
// It provides methods for creating, retrieving, updating, deleting, and listing indexes.
// We will not reinvent the wheel and use the resource.Access type from the cloud-native-utils package.
type IndexRepository resource.Access[IndexID, Index]
