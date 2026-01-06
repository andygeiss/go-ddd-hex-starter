package indexing_test

import (
	"encoding/json"
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_EventFileIndexCreated_NewEventFileIndexCreated_Should_Create_Empty_Event(t *testing.T) {
	// Arrange & Act
	event := indexing.NewEventFileIndexCreated()

	// Assert
	assert.That(t, "IndexID must be empty", string(event.IndexID), "")
	assert.That(t, "FileCount must be zero", event.FileCount, 0)
}

func Test_EventFileIndexCreated_WithIndexID_Should_Set_IndexID(t *testing.T) {
	// Arrange
	expectedID := indexing.IndexID("test-index-123")

	// Act
	event := indexing.NewEventFileIndexCreated().WithIndexID(expectedID)

	// Assert
	assert.That(t, "IndexID must match", event.IndexID, expectedID)
}

func Test_EventFileIndexCreated_WithFileCount_Should_Set_FileCount(t *testing.T) {
	// Arrange
	expectedCount := 42

	// Act
	event := indexing.NewEventFileIndexCreated().WithFileCount(expectedCount)

	// Assert
	assert.That(t, "FileCount must match", event.FileCount, expectedCount)
}

func Test_EventFileIndexCreated_Topic_Should_Return_Correct_Topic(t *testing.T) {
	// Arrange
	event := indexing.NewEventFileIndexCreated()

	// Act
	topic := event.Topic()

	// Assert
	assert.That(t, "Topic must be correct", topic, indexing.EventTopicFileIndexCreated)
	assert.That(t, "Topic constant must have expected value", topic, "indexing.file_index_created")
}

func Test_EventFileIndexCreated_Builder_Chain_Should_Set_All_Fields(t *testing.T) {
	// Arrange
	expectedID := indexing.IndexID("chain-test-id")
	expectedCount := 100

	// Act
	event := indexing.NewEventFileIndexCreated().
		WithIndexID(expectedID).
		WithFileCount(expectedCount)

	// Assert
	assert.That(t, "IndexID must match", event.IndexID, expectedID)
	assert.That(t, "FileCount must match", event.FileCount, expectedCount)
}

func Test_EventFileIndexCreated_JSON_Serialization_Should_Work(t *testing.T) {
	// Arrange
	event := indexing.NewEventFileIndexCreated().
		WithIndexID("json-test-id").
		WithFileCount(50)

	// Act
	data, err := json.Marshal(event)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "JSON must contain index_id", containsString(string(data), `"index_id":"json-test-id"`), true)
	assert.That(t, "JSON must contain file_count", containsString(string(data), `"file_count":50`), true)
}

func Test_EventFileIndexCreated_JSON_Deserialization_Should_Work(t *testing.T) {
	// Arrange
	jsonData := `{"index_id":"deserialize-test","file_count":25}`
	event := indexing.NewEventFileIndexCreated()

	// Act
	err := json.Unmarshal([]byte(jsonData), event)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "IndexID must match", string(event.IndexID), "deserialize-test")
	assert.That(t, "FileCount must match", event.FileCount, 25)
}

// containsString is a helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
