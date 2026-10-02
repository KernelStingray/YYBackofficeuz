package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client is a wrapper around the amqp connection and channel
type Client struct {
	conn            *amqp.Connection
	ch              *amqp.Channel
	url             string
	mu              sync.RWMutex
	isConnected     bool
	ready           chan struct{}
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	done            chan bool
}

// NewClient initializes a RabbitMQ client with auto-reconnect
func NewClient(url string) *Client {
	c := &Client{
		url:   url,
		done:  make(chan bool),
		ready: make(chan struct{}),
	}
	go c.handleReconnect()
	return c
}

// handleReconnect manages connection lifecycle and reconnects on failure
func (c *Client) handleReconnect() {
	for {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()

		log.Println("RabbitMQ: Attempting to connect...")

		conn, err := amqp.Dial(c.url)
		if err != nil {
			log.Printf("RabbitMQ: Failed to connect: %v. Retrying in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			log.Printf("RabbitMQ: Failed to open channel: %v. Retrying in 3s...", err)
			conn.Close()
			time.Sleep(3 * time.Second)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.ch = ch
		c.notifyConnClose = make(chan *amqp.Error, 1)
		c.notifyChanClose = make(chan *amqp.Error, 1)
		c.conn.NotifyClose(c.notifyConnClose)
		c.ch.NotifyClose(c.notifyChanClose)
		c.isConnected = true

		// Signal all waiting callers that connection is ready
		select {
		case <-c.ready:
		default:
			close(c.ready)
		}
		c.mu.Unlock()

		log.Println("RabbitMQ: Successfully connected!")

		select {
		case <-c.done:
			log.Println("RabbitMQ: Closing reconnect loop.")
			return
		case err := <-c.notifyConnClose:
			log.Printf("RabbitMQ: Connection closed: %v", err)
		case err := <-c.notifyChanClose:
			log.Printf("RabbitMQ: Channel closed: %v", err)
		}

		c.mu.Lock()
		c.ready = make(chan struct{})
		c.mu.Unlock()
	}
}

// WaitForConnection blocks until RabbitMQ connection is established or timeout/context expires
func (c *Client) WaitForConnection(ctx context.Context, timeout time.Duration) error {
	c.mu.RLock()
	if c.isConnected {
		c.mu.RUnlock()
		return nil
	}
	ready := c.ready
	c.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("rabbitmq: connection wait timeout: %w", ctx.Err())
		case <-c.done:
			return fmt.Errorf("rabbitmq: client closed")
		case <-ready:
			c.mu.RLock()
			connected := c.isConnected
			c.mu.RUnlock()
			if connected {
				return nil
			}
			c.mu.RLock()
			ready = c.ready
			c.mu.RUnlock()
		}
	}
}

// Publish sends a message to the specified exchange and routing key
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	if err := c.WaitForConnection(ctx, 10*time.Second); err != nil {
		return err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.ch == nil {
		return fmt.Errorf("rabbitmq: not connected")
	}

	err := c.ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		})
	return err
}

// Consume starts a consumer on the given queue and returns a channel of deliveries
func (c *Client) Consume(queueName string) (<-chan amqp.Delivery, error) {
	if err := c.WaitForConnection(context.Background(), 15*time.Second); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.ch == nil {
		return nil, fmt.Errorf("rabbitmq: not connected")
	}

	msgs, err := c.ch.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// SetupExchangeQueue declares an exchange and queue, and binds them
func (c *Client) SetupExchangeQueue(exchange, exchangeType, queueName, routingKey string) error {
	if err := c.WaitForConnection(context.Background(), 15*time.Second); err != nil {
		return err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.ch == nil {
		return fmt.Errorf("rabbitmq: not connected")
	}

	// Declare exchange
	err := c.ch.ExchangeDeclare(
		exchange,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	// Declare queue
	q, err := c.ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// Bind queue
	err = c.ch.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	return nil
}

// SetupExchangeQueueWithDLQ declares an exchange and queue with DLQ settings, and binds them
func (c *Client) SetupExchangeQueueWithDLQ(exchange, exchangeType, queueName, routingKey, dlx, dlqRoutingKey string) error {
	if err := c.WaitForConnection(context.Background(), 15*time.Second); err != nil {
		return err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.ch == nil {
		return fmt.Errorf("rabbitmq: not connected")
	}

	// Declare exchange
	err := c.ch.ExchangeDeclare(
		exchange,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	// Declare DLX
	err = c.ch.ExchangeDeclare(
		dlx,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dlx: %v", err)
	}

	// Declare Queue with DLX arguments
	args := amqp.Table{
		"x-dead-letter-exchange":    dlx,
		"x-dead-letter-routing-key": dlqRoutingKey,
	}
	q, err := c.ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// Bind queue
	err = c.ch.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	// Declare DLQ
	dlq, err := c.ch.QueueDeclare(
		queueName+".dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dlq: %v", err)
	}

	// Bind DLQ
	err = c.ch.QueueBind(
		dlq.Name,
		dlqRoutingKey,
		dlx,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dlq: %v", err)
	}

	return nil
}

// Close closes the connection gracefully
func (c *Client) Close() {
	close(c.done)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
