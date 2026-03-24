package contracts

// MessageBrokerAdapter composes BrokerPublisher and BrokerSubscriber into a
// single swappable Strategy for broker backends (Kafka, RabbitMQ, NATS, etc.).
// Implementations can be registered at startup and injected wherever
// event-driven correlation or outbound publishing is required.
type MessageBrokerAdapter interface {
	BrokerPublisher
	BrokerSubscriber
}
