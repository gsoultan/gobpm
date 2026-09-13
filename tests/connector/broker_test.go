package connector_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	serviceimpl "github.com/gsoultan/metis/server/domains/services/impl"
	"github.com/gsoultan/metis/server/repositories"
	"github.com/gsoultan/metis/tests/testutils"
)

// The RabbitMQ connector, against a real broker.
//
// Everything else in this package proves the connector survives a bad payload
// and refuses an incomplete configuration. None of it proves a message is ever
// delivered — the executor dials a broker, and with no broker to dial the only
// outcome a test could observe was the dial failing. So the README's claim of
// "production-ready messaging" rested on nothing, which is the kind of gap that
// is found by a partner rather than by CI.
//
// The suite is gated on a broker URL and CI fails on any skipped test, so
// adding it without adding the service to the workflow breaks the build rather
// than quietly running nothing.

const brokerURLEnv = "METIS_TEST_RABBITMQ_URL"

func brokerURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv(brokerURLEnv)
	if url == "" {
		t.Skipf("set %s to run this against a live RabbitMQ broker", brokerURLEnv)
	}
	return url
}

// consumeOne declares a temporary queue bound to nothing, and returns a
// function that waits for one message on it.
func consumeOne(t *testing.T, url, queue string) func() []byte {
	t.Helper()

	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatalf("dial the broker: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("open a channel: %v", err)
	}
	t.Cleanup(func() { _ = ch.Close() })

	if _, err := ch.QueueDeclare(queue, false, true, false, false, nil); err != nil {
		t.Fatalf("declare %s: %v", queue, err)
	}
	deliveries, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("consume %s: %v", queue, err)
	}

	return func() []byte {
		select {
		case d := <-deliveries:
			return d.Body
		case <-time.After(5 * time.Second):
			t.Fatalf("nothing arrived on %s within 5s", queue)
			return nil
		}
	}
}

// TestPublishingReachesTheQueue is the claim the README makes, asserted for the
// first time: a service task wired to this connector puts the message on the
// broker, and what arrives is the payload the process built.
func TestPublishingReachesTheQueue(t *testing.T) {
	url := brokerURL(t)
	const queue = "metis-test-delivery"

	await := consumeOne(t, url, queue)

	svc := serviceimpl.NewConnectorService(repositories.NewRepository(testutils.SetupTestConn(t)))
	result, err := svc.ExecuteConnector(context.Background(), "rabbitmq-publish",
		map[string]any{"url": url, "queue": queue},
		map[string]any{"claim_id": "C-42", "amount": 1500})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if result["status"] != "published" {
		t.Errorf("status: got %v, want published", result["status"])
	}

	var got map[string]any
	if err := json.Unmarshal(await(), &got); err != nil {
		t.Fatalf("the delivered body is not the JSON payload: %v", err)
	}
	if got["claim_id"] != "C-42" {
		t.Errorf("claim_id arrived as %v, want C-42", got["claim_id"])
	}
}

// TestPublishingToNowhereIsNotReportedAsSent is the one that matters.
//
// AMQP publishes are fire-and-forget. Without publisher confirms a broker
// accepts a message and says nothing, and without the mandatory flag it discards
// one it cannot route — to an exchange that does not exist, or with a routing
// key nothing is bound to. Both look identical to a successful send.
//
// In an engine that runs other people's obligations that is the worst available
// outcome: the service task reports success, the token advances, the audit trail
// records a message that was sent, and the message is gone. There is nothing to
// investigate because nothing looks wrong.
func TestPublishingToNowhereIsNotReportedAsSent(t *testing.T) {
	url := brokerURL(t)

	svc := serviceimpl.NewConnectorService(repositories.NewRepository(testutils.SetupTestConn(t)))
	_, err := svc.ExecuteConnector(context.Background(), "rabbitmq-publish",
		map[string]any{
			"url":         url,
			"routing_key": "metis-test-nothing-is-bound-to-this",
		},
		map[string]any{"claim_id": "C-43"})

	if err == nil {
		t.Fatal("a message that the broker could not route was reported as published")
	}
}

// TestPublishingToAnExchangeThatDoesNotExistIsRefused covers the other way a
// message disappears: naming an exchange that was never declared.
func TestPublishingToAnExchangeThatDoesNotExistIsRefused(t *testing.T) {
	url := brokerURL(t)

	svc := serviceimpl.NewConnectorService(repositories.NewRepository(testutils.SetupTestConn(t)))
	_, err := svc.ExecuteConnector(context.Background(), "rabbitmq-publish",
		map[string]any{
			"url":         url,
			"exchange":    "metis-test-no-such-exchange",
			"routing_key": "anything",
		},
		map[string]any{"claim_id": "C-44"})

	if err == nil {
		t.Fatal("a message addressed to an exchange that does not exist was reported as published")
	}
}

// TestPublishingThroughAnExchangeReachesTheBoundQueue covers the other route.
//
// The fix derives a routing key from the queue name only when no exchange is
// named; this proves it did not break the case where one is, which is the
// configuration the connector's description actually advertises.
func TestPublishingThroughAnExchangeReachesTheBoundQueue(t *testing.T) {
	url := brokerURL(t)
	const (
		exchange   = "metis-test-exchange"
		queue      = "metis-test-bound"
		routingKey = "claims.settled"
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatalf("dial the broker: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("open a channel: %v", err)
	}
	t.Cleanup(func() { _ = ch.Close() })

	if err := ch.ExchangeDeclare(exchange, "direct", false, true, false, false, nil); err != nil {
		t.Fatalf("declare the exchange: %v", err)
	}
	if _, err := ch.QueueDeclare(queue, false, true, false, false, nil); err != nil {
		t.Fatalf("declare the queue: %v", err)
	}
	if err := ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		t.Fatalf("bind the queue: %v", err)
	}
	deliveries, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}

	svc := serviceimpl.NewConnectorService(repositories.NewRepository(testutils.SetupTestConn(t)))
	if _, err := svc.ExecuteConnector(context.Background(), "rabbitmq-publish",
		map[string]any{"url": url, "exchange": exchange, "routing_key": routingKey},
		map[string]any{"claim_id": "C-45"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case d := <-deliveries:
		var got map[string]any
		if err := json.Unmarshal(d.Body, &got); err != nil {
			t.Fatalf("the delivered body is not the JSON payload: %v", err)
		}
		if got["claim_id"] != "C-45" {
			t.Errorf("claim_id arrived as %v, want C-45", got["claim_id"])
		}
	case <-time.After(5 * time.Second):
		t.Fatal("nothing arrived on the bound queue within 5s")
	}
}
