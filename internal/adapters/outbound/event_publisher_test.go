package outbound_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/hotel-booking/internal/adapters/outbound"
)

// ============================================================================
// EventPublisher Tests
// ============================================================================

type mockDispatcher struct {
	publishErr        error
	subscribeErr      error
	publishedMessages []messaging.Message
}

func (m *mockDispatcher) Publish(ctx context.Context, msg messaging.Message) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedMessages = append(m.publishedMessages, msg)
	return nil
}

func (m *mockDispatcher) Subscribe(ctx context.Context, topic string, handler service.Function[messaging.Message, messaging.MessageState]) error {
	return m.subscribeErr
}

type testEvent struct {
	EventTopic string `json:"topic"`
	Data       string `json:"data"`
}

func (e *testEvent) Topic() string {
	return e.EventTopic
}

func Test_EventPublisher_Publish_Should_Succeed(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	ctx := context.Background()

	event := &testEvent{
		EventTopic: "test.topic",
		Data:       "test data",
	}

	// Act
	err := publisher.Publish(ctx, event)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 published message", len(dispatcher.publishedMessages), 1)
	assert.That(t, "topic must match", dispatcher.publishedMessages[0].Topic, "test.topic")

	var decoded testEvent
	unmarshalErr := json.Unmarshal(dispatcher.publishedMessages[0].Data, &decoded)
	assert.That(t, "unmarshal must succeed", unmarshalErr == nil, true)
	assert.That(t, "data must match", decoded.Data, "test data")
}

func Test_EventPublisher_Publish_Dispatcher_Error_Should_Return_Error(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{
		publishErr: errors.New("dispatcher error"),
	}
	publisher := outbound.NewEventPublisher(dispatcher)
	ctx := context.Background()

	event := &testEvent{
		EventTopic: "test.topic",
		Data:       "test data",
	}

	// Act
	err := publisher.Publish(ctx, event)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_EventPublisher_Publish_Multiple_Events_Should_Succeed(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	ctx := context.Background()

	event1 := &testEvent{EventTopic: "topic1", Data: "data1"}
	event2 := &testEvent{EventTopic: "topic2", Data: "data2"}
	event3 := &testEvent{EventTopic: "topic3", Data: "data3"}

	// Act
	_ = publisher.Publish(ctx, event1)
	_ = publisher.Publish(ctx, event2)
	_ = publisher.Publish(ctx, event3)

	// Assert
	assert.That(t, "must have 3 published messages", len(dispatcher.publishedMessages), 3)
}

func Test_EventPublisher_Publish_Different_Topics_Should_Use_Correct_Topic(t *testing.T) {
	// Arrange
	dispatcher := &mockDispatcher{}
	publisher := outbound.NewEventPublisher(dispatcher)
	ctx := context.Background()

	event1 := &testEvent{EventTopic: "reservation.created", Data: "res1"}
	event2 := &testEvent{EventTopic: "payment.authorized", Data: "pay1"}

	// Act
	_ = publisher.Publish(ctx, event1)
	_ = publisher.Publish(ctx, event2)

	// Assert
	assert.That(t, "first message topic must match", dispatcher.publishedMessages[0].Topic, "reservation.created")
	assert.That(t, "second message topic must match", dispatcher.publishedMessages[1].Topic, "payment.authorized")
}
