package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/services/contracts"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	workerID          = "messaging-bridge"
	maxTasks          = 10
	lockDurationSec   = 30
	pollInterval      = 5 * time.Second
	reconnectInterval = 5 * time.Second
)

type messagingService struct {
	engine      contracts.ExecutionEngine
	externalSvc contracts.ExternalTaskService
	cancels     sync.Map // string -> context.CancelFunc
	wg          sync.WaitGroup
}

func NewMessagingService(engine contracts.ExecutionEngine, externalSvc contracts.ExternalTaskService) contracts.MessagingService {
	return &messagingService{
		engine:      engine,
		externalSvc: externalSvc,
	}
}

func (s *messagingService) StartBridge(ctx context.Context, projectID uuid.UUID, topic string, rabbitURL string, exchange string, routingKey string) error {
	id := fmt.Sprintf("bridge-%s-%s", projectID, topic)
	if _, loaded := s.cancels.Load(id); loaded {
		return fmt.Errorf("bridge for topic %s already running", topic)
	}

	childCtx, cancel := context.WithCancel(ctx)
	s.cancels.Store(id, cancel)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.cancels.Delete(id)
		s.runBridge(childCtx, projectID, topic, rabbitURL, exchange, routingKey)
	}()

	return nil
}

func (s *messagingService) runBridge(ctx context.Context, projectID uuid.UUID, topic string, rabbitURL string, exchange string, routingKey string) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var conn *amqp.Connection
	var ch *amqp.Channel
	var err error

	cleanup := func() {
		if ch != nil {
			ch.Close()
		}
		if conn != nil {
			conn.Close()
		}
		ch = nil
		conn = nil
	}
	defer cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if conn == nil || conn.IsClosed() {
				cleanup()
				conn, err = amqp.Dial(rabbitURL)
				if err != nil {
					log.Error().Err(err).Msg("Bridge RabbitMQ connection error")
					continue
				}
				ch, err = conn.Channel()
				if err != nil {
					log.Error().Err(err).Msg("Bridge RabbitMQ channel error")
					cleanup()
					continue
				}
			}

			// Fetch and lock tasks
			tasks, err := s.externalSvc.FetchAndLock(ctx, topic, workerID, maxTasks, lockDurationSec)
			if err != nil {
				log.Error().Err(err).Msg("Bridge fetch error")
				continue
			}

			for _, task := range tasks {
				body, _ := json.Marshal(task)
				err = ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
					Headers: amqp.Table{
						"task_id": task.ID.String(),
					},
				})
				if err != nil {
					log.Error().Err(err).Msg("Bridge publish error")
					// Task will timeout and be retried
					continue
				}
				log.Info().Str("taskID", task.ID.String()).Msg("Forwarded external task to RabbitMQ")
			}
		}
	}
}

func (s *messagingService) StartInboundConsumer(ctx context.Context, projectID uuid.UUID, rabbitURL string, queueName string, messageName string) error {
	id := fmt.Sprintf("consumer-%s-%s", projectID, queueName)
	if _, loaded := s.cancels.Load(id); loaded {
		return fmt.Errorf("consumer for queue %s already running", queueName)
	}

	childCtx, cancel := context.WithCancel(ctx)
	s.cancels.Store(id, cancel)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.cancels.Delete(id)
		s.runConsumer(childCtx, projectID, rabbitURL, queueName, messageName)
	}()

	return nil
}

func (s *messagingService) runConsumer(ctx context.Context, projectID uuid.UUID, rabbitURL string, queueName string, messageName string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := s.consumeOnce(ctx, projectID, rabbitURL, queueName, messageName)
			if err != nil {
				log.Error().Err(err).Dur("retryIn", reconnectInterval).Msg("Consumer error; retrying")
				select {
				case <-ctx.Done():
					return
				case <-time.After(reconnectInterval):
				}
			}
		}
	}
}

func (s *messagingService) consumeOnce(ctx context.Context, projectID uuid.UUID, rabbitURL string, queueName string, messageName string) error {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.ConsumeWithContext(ctx, q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range msgs {
		var payload map[string]any
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal inbound message")
			continue
		}

		correlationKey, _ := payload["correlation_key"].(string)
		if correlationKey == "" {
			// Try to get from headers
			if v, ok := d.Headers["correlation_key"].(string); ok {
				correlationKey = v
			}
		}

		log.Info().Str("messageName", messageName).Str("correlationKey", correlationKey).Msg("Received inbound message")
		err = s.engine.SendMessage(ctx, projectID, messageName, correlationKey, payload)
		if err != nil {
			log.Error().Err(err).Msg("Error sending inbound message to engine")
		}
	}

	return nil
}

func (s *messagingService) StopAll() {
	s.cancels.Range(func(key, value any) bool {
		cancel := value.(context.CancelFunc)
		cancel()
		return true
	})
	s.wg.Wait()
}
