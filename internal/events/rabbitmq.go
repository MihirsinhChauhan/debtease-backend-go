package events

import (
	"DebtEase/internal/logger"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// Exchange and queue names
	exchangeName = "notification.events"
	queueName    = "notification.queue"
	routingKey   = "notification.#" // Wildcard for all notification events
)

var (
	conn *amqp.Connection
	ch *amqp.Channel
	connMutex sync.RWMutex
	chMutex sync.RWMutex
)


// GetConnection returns a singleton RabbitMQ connection
func GetConnection(rabbitmqURL string) (*amqp.Connection, error) {
	connMutex.RLock()
	if conn != nil && !conn.IsClosed() {
		connMutex.RUnlock()
		return conn, nil
	}
	connMutex.RUnlock()

	connMutex.Lock()
	defer connMutex.Unlock()

	// Double-check after acquiring write lock
	if conn != nil && !conn.IsClosed() {
		return conn, nil
	}

	newConn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	conn = newConn
	logger.Logger.Infow("RabbitMQ connection established")
	return conn, nil
}

// GetChannel returns a singleton RabbitMQ channel
func GetChannel(rabbitmqURL string) (*amqp.Channel, error) {
	chMutex.RLock()
	if ch != nil && !ch.IsClosed() {
		chMutex.RUnlock()
		return ch, nil
	}
	chMutex.RUnlock()

	chMutex.Lock()
	defer chMutex.Unlock()

	// Double-check after acquiring write lock
	if ch != nil && !ch.IsClosed() {
		return ch, nil
	}

	conn, err := GetConnection(rabbitmqURL)
	if err != nil {
		return nil, err
	}

	newCh, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (topic exchange for routing by event type)
	err = newCh.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		newCh.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = newCh.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		newCh.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = newCh.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		newCh.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	ch = newCh
	logger.Logger.Infow("RabbitMQ channel established", "exchange", exchangeName, "queue", queueName)
	return ch, nil
}

// PublishEvent publishes an event to RabbitMQ
func PublishEvent(ctx context.Context, rabbitmqURL string, event interface{}) error {
	channel, err := GetChannel(rabbitmqURL)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Marshal event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Determine routing key from event type
	var routingKey string

	switch event.(type) {
	case *UserRegisteredEvent:
		routingKey = "notification.user.registered"
	default:
		// Try to extract event_type from BaseEvent if it exists
		if baseEvent, ok := event.(interface{ GetEventType() string }); ok {
			routingKey = fmt.Sprintf("notification.%s", baseEvent.GetEventType())
		} else {
			routingKey = "notification.unknown"
		}
	}

	// Publish message
	err = channel.PublishWithContext(
		ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Make message persistent
			Timestamp:    time.Now(),
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	logger.Logger.Infow("Event published to RabbitMQ",
		"routing_key", routingKey,
		"event_size", len(body),
	)

	return nil
}


// CloseConnections closes RabbitMQ connections (for graceful shutdown)
func CloseConnections() {
	chMutex.Lock()
	if ch != nil && !ch.IsClosed() {
		ch.Close()
		ch = nil
	}
	chMutex.Unlock()

	connMutex.Lock()
	if conn != nil && !conn.IsClosed() {
		conn.Close()
		conn = nil
	}
	connMutex.Unlock()

	logger.Logger.Info("RabbitMQ connections closed")
}