package indexing

// This file contains the event types for the indexing domain.

const (
	EventTopicFileIndexCreated = "indexing.file_index_created"
)

// EventFileIndexCreated represents a file index created event.
type EventFileIndexCreated struct {
	IndexID   IndexID `json:"index_id"`
	FileCount int     `json:"file_count"`
}

// NewEventFileIndexCreated creates a new EventFileIndexCreated instance.
// Use the builder methods to set the fields.
func NewEventFileIndexCreated() *EventFileIndexCreated {
	return &EventFileIndexCreated{}
}

// WithIndexID sets the IndexID field.
func (e *EventFileIndexCreated) WithIndexID(id IndexID) *EventFileIndexCreated {
	e.IndexID = id
	return e
}

// WithFileCount sets the FileCount field.
func (e *EventFileIndexCreated) WithFileCount(count int) *EventFileIndexCreated {
	e.FileCount = count
	return e
}

// Topic returns the topic for the event.
func (e *EventFileIndexCreated) Topic() string {
	return EventTopicFileIndexCreated
}
