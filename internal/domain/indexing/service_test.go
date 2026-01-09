package indexing_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// Tests follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library.

// --- EventFileIndexCreated tests ---

func Test_EventFileIndexCreated_Builder_Chain_Should_Set_All_Fields(t *testing.T) {
	// Arrange
	expectedID := indexing.IndexID("chain-test-id")
	expectedCount := 100

	// Act
	evt := indexing.NewEventFileIndexCreated().
		WithIndexID(expectedID).
		WithFileCount(expectedCount)

	// Assert
	assert.That(t, "IndexID must match", evt.IndexID, expectedID)
	assert.That(t, "FileCount must match", evt.FileCount, expectedCount)
}

func Test_EventFileIndexCreated_JSON_Deserialization_Should_Work(t *testing.T) {
	// Arrange
	jsonData := `{"index_id":"deserialize-test","file_count":25}`
	evt := indexing.NewEventFileIndexCreated()

	// Act
	err := json.Unmarshal([]byte(jsonData), evt)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "IndexID must match", string(evt.IndexID), "deserialize-test")
	assert.That(t, "FileCount must match", evt.FileCount, 25)
}

func Test_EventFileIndexCreated_JSON_Serialization_Should_Work(t *testing.T) {
	// Arrange
	evt := indexing.NewEventFileIndexCreated().
		WithIndexID("json-test-id").
		WithFileCount(50)

	// Act
	data, err := json.Marshal(evt)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "JSON must contain index_id", containsString(string(data), `"index_id":"json-test-id"`), true)
	assert.That(t, "JSON must contain file_count", containsString(string(data), `"file_count":50`), true)
}

func Test_EventFileIndexCreated_NewEventFileIndexCreated_Should_Create_Empty_Event(t *testing.T) {
	// Arrange & Act
	evt := indexing.NewEventFileIndexCreated()

	// Assert
	assert.That(t, "IndexID must be empty", string(evt.IndexID), "")
	assert.That(t, "FileCount must be zero", evt.FileCount, 0)
}

func Test_EventFileIndexCreated_Topic_Should_Return_Correct_Topic(t *testing.T) {
	// Arrange
	evt := indexing.NewEventFileIndexCreated()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "Topic must be correct", topic, indexing.EventTopicFileIndexCreated)
	assert.That(t, "Topic constant must have expected value", topic, "indexing.file_index_created")
}

func Test_EventFileIndexCreated_WithFileCount_Should_Set_FileCount(t *testing.T) {
	// Arrange
	expectedCount := 42

	// Act
	evt := indexing.NewEventFileIndexCreated().WithFileCount(expectedCount)

	// Assert
	assert.That(t, "FileCount must match", evt.FileCount, expectedCount)
}

func Test_EventFileIndexCreated_WithIndexID_Should_Set_IndexID(t *testing.T) {
	// Arrange
	expectedID := indexing.IndexID("test-index-123")

	// Act
	evt := indexing.NewEventFileIndexCreated().WithIndexID(expectedID)

	// Assert
	assert.That(t, "IndexID must match", evt.IndexID, expectedID)
}

// --- IndexingService tests ---

func Test_IndexingService_CreateIndex_With_Mockup_Should_Be_Called(t *testing.T) {
	// Arrange
	sut, publisher := setupIndexingService()
	publisher = publisher.(*mockEventPublisher)
	path := "testdata/index.json"
	ctx := context.Background()

	// Act
	err := sut.CreateIndex(ctx, path)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "publisher must be called", publisher.(*mockEventPublisher).Published, true)
}

func Test_IndexingService_CreateIndex_With_Mockup_Should_Return_Two_Entries(t *testing.T) {
	// Arrange
	sut, _ := setupIndexingService()
	path := "testdata/index.json"
	ctx := context.Background()

	// Act
	err := sut.CreateIndex(ctx, path)
	files, err2 := sut.IndexFiles(ctx, path)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "err2 must be nil", err2 == nil, true)
	assert.That(t, "index must have two entries", len(files) == 2, true)
}

// --- Test helpers ---

// containsString is a helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockEventPublisher is a simple mock implementation for the EventPublisher interface.
type mockEventPublisher struct {
	Published bool
}

// Publish publishes the message.
func (a *mockEventPublisher) Publish(_ context.Context, _ event.Event) error {
	a.Published = true
	return nil
}

// mockFileReader is a simple mock implementation for the FileReader interface.
type mockFileReader struct {
	fileInfos []indexing.FileInfo
}

// ReadFileInfos returns the slice of file infos.
func (a *mockFileReader) ReadFileInfos(_ context.Context, _ string) ([]indexing.FileInfo, error) {
	return a.fileInfos, nil
}

// setupIndexingService creates a new IndexingService with mocked dependencies.
func setupIndexingService() (*indexing.IndexingService, event.EventPublisher) {
	mockFileReader := &mockFileReader{
		fileInfos: []indexing.FileInfo{
			{AbsPath: "test/path/file1.txt", Size: 100},
			{AbsPath: "test/path/file2.txt", Size: 200},
		},
	}
	mockIndexRepository := resource.NewMockAccess[indexing.IndexID, indexing.Index]()
	mockIndexRepository.WithCreateFn(
		func(_ context.Context, _ indexing.IndexID, _ indexing.Index) error {
			return nil
		})
	mockEventPublisher := &mockEventPublisher{}
	service := indexing.NewIndexingService(mockFileReader, mockIndexRepository, mockEventPublisher)
	return service, mockEventPublisher
}
