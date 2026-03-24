package entities

import "time"

// BrokerMessage represents an event message exchanged through an external
// message broker (Kafka, RabbitMQ, NATS, etc.).
// It carries correlation data so the engine can match the message to a
// waiting IntermediateCatchEvent token.
type BrokerMessage struct {
	// ID is a unique message identifier assigned by the producer.
	ID string
	// Topic is the broker topic or queue name the message was published to.
	Topic string
	// CorrelationKey is matched against process instance variables to find
	// the correct waiting token (e.g., order ID, customer ID).
	CorrelationKey string
	// Payload contains the message data made available as process variables.
	Payload map[string]any
	// PublishedAt is when the message was originally produced.
	PublishedAt time.Time
}
