package event_test

import (
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_Event_Interface_With_EventFileIndexCreated_Should_Return_Topic(t *testing.T) {
	// Arrange
	var e event.Event = indexing.NewEventFileIndexCreated()

	// Act
	topic := e.Topic()

	// Assert
	assert.That(t, "topic must be 'indexing.file_index_created'", topic, indexing.EventTopicFileIndexCreated)
}

func Test_EventFactoryFn_With_EventFileIndexCreated_Should_Create_Event(t *testing.T) {
	// Arrange
	factory := func() event.Event {
		return indexing.NewEventFileIndexCreated()
	}

	// Act
	e := factory()

	// Assert
	assert.That(t, "event must not be nil", e != nil, true)
	assert.That(t, "event topic must be correct", e.Topic(), indexing.EventTopicFileIndexCreated)
}

func Test_EventHandlerFn_With_Event_Should_Handle_Without_Error(t *testing.T) {
	// Arrange
	handled := false
	var handler event.EventHandlerFn = func(_ event.Event) error {
		handled = true
		return nil
	}
	e := indexing.NewEventFileIndexCreated()

	// Act
	err := handler(e)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "event must be handled", handled, true)
}
