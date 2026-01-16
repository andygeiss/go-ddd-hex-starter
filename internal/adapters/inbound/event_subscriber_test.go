package inbound_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/cloud-native-utils/service"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// MessageFunction type alias for cleaner code.
type MessageFunction = service.Function[messaging.Message, messaging.MessageState]

// Mock dispatcher for testing.
type mockDispatcher struct {
	subscribedTopics  []string
	subscribeErr      error
	messageHandler    MessageFunction
	publishedMessages []messaging.Message
}

func (m *mockDispatcher) Publish(ctx context.Context, msg messaging.Message) error {
	m.publishedMessages = append(m.publishedMessages, msg)
	return nil
}

func (m *mockDispatcher) Subscribe(ctx context.Context, topic string, handler MessageFunction) error {
	if m.subscribeErr != nil {
		return m.subscribeErr
	}
	m.subscribedTopics = append(m.subscribedTopics, topic)
	m.messageHandler = handler
	return nil
}

// Test event for testing.
type testEvent struct {
	EventTopic string `json:"topic"`
	Data       string `json:"data"`
}

func (e *testEvent) Topic() string {
	return e.EventTopic
}

func newTestEvent() event.Event {
	return &testEvent{}
}

func Test_EventSubscriber_Subscribe_Should_Register_Topic(t *testing.T) {
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	ctx := context.Background()

	handler := func(e event.Event) error {
		return nil
	}

	err := subscriber.Subscribe(ctx, "test.topic", newTestEvent, handler)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(dispatcher.subscribedTopics) != 1 {
		t.Errorf("expected 1 subscribed topic, got %d", len(dispatcher.subscribedTopics))
	}

	if dispatcher.subscribedTopics[0] != "test.topic" {
		t.Errorf("expected topic 'test.topic', got %s", dispatcher.subscribedTopics[0])
	}
}

func Test_EventSubscriber_Subscribe_Dispatcher_Error_Should_Return_Error(t *testing.T) {
	dispatcher := &mockDispatcher{
		subscribeErr: errors.New("subscribe error"),
	}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	ctx := context.Background()

	handler := func(e event.Event) error {
		return nil
	}

	err := subscriber.Subscribe(ctx, "test.topic", newTestEvent, handler)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func Test_EventSubscriber_Message_Handler_Should_Call_Event_Handler(t *testing.T) {
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	ctx := context.Background()

	var receivedEvent *testEvent
	handler := func(e event.Event) error {
		receivedEvent = e.(*testEvent)
		return nil
	}

	_ = subscriber.Subscribe(ctx, "test.topic", newTestEvent, handler)

	// Simulate receiving a message.
	eventData := &testEvent{
		EventTopic: "test.topic",
		Data:       "test message data",
	}
	encoded, marshalErr := json.Marshal(eventData)
	if marshalErr != nil {
		t.Fatalf("failed to marshal event data: %v", marshalErr)
	}
	msg := messaging.NewMessage("test.topic", encoded)

	state, err := dispatcher.messageHandler(ctx, msg)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if state != messaging.MessageStateCompleted {
		t.Errorf("expected state Completed, got %v", state)
	}

	if receivedEvent == nil {
		t.Fatal("expected handler to be called with event")
	}

	if receivedEvent.Data != "test message data" {
		t.Errorf("expected event data 'test message data', got %s", receivedEvent.Data)
	}
}

func Test_EventSubscriber_Message_Handler_Invalid_JSON_Should_Return_Failed(t *testing.T) {
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	ctx := context.Background()

	handler := func(e event.Event) error {
		return nil
	}

	_ = subscriber.Subscribe(ctx, "test.topic", newTestEvent, handler)

	// Simulate receiving invalid JSON.
	msg := messaging.NewMessage("test.topic", []byte("invalid json"))

	state, err := dispatcher.messageHandler(ctx, msg)

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	if state != messaging.MessageStateFailed {
		t.Errorf("expected state Failed, got %v", state)
	}
}

func Test_EventSubscriber_Message_Handler_Error_Should_Return_Failed(t *testing.T) {
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	ctx := context.Background()

	handler := func(e event.Event) error {
		return errors.New("handler error")
	}

	_ = subscriber.Subscribe(ctx, "test.topic", newTestEvent, handler)

	// Simulate receiving a valid message.
	eventData := &testEvent{EventTopic: "test.topic", Data: "data"}
	encoded, marshalErr := json.Marshal(eventData)
	if marshalErr != nil {
		t.Fatalf("failed to marshal event data: %v", marshalErr)
	}
	msg := messaging.NewMessage("test.topic", encoded)

	state, err := dispatcher.messageHandler(ctx, msg)

	if err == nil {
		t.Error("expected error from handler, got nil")
	}

	if state != messaging.MessageStateFailed {
		t.Errorf("expected state Failed, got %v", state)
	}
}

func Test_EventSubscriber_Subscribe_Multiple_Topics_Should_Succeed(t *testing.T) {
	dispatcher := &mockDispatcher{}
	subscriber := inbound.NewEventSubscriber(dispatcher)
	ctx := context.Background()

	handler := func(e event.Event) error {
		return nil
	}

	_ = subscriber.Subscribe(ctx, "topic1", newTestEvent, handler)
	_ = subscriber.Subscribe(ctx, "topic2", newTestEvent, handler)
	_ = subscriber.Subscribe(ctx, "topic3", newTestEvent, handler)

	if len(dispatcher.subscribedTopics) != 3 {
		t.Errorf("expected 3 subscribed topics, got %d", len(dispatcher.subscribedTopics))
	}
}
