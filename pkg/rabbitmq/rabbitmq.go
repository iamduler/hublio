package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type rabbitMQService struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	logger  *zerolog.Logger
}

func NewRabbitMQService(amqpURL string, logger *zerolog.Logger) (RabbitMQService, error) {
	// Connect to RabbitMQ
	conn, err := amqp091.Dial(amqpURL)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to RabbitMQ")
		return nil, err
	}

	// Create a channel
	ch, err := conn.Channel()

	if err != nil {
		conn.Close()
		logger.Error().Err(err).Msg("Failed to create channel")
		return nil, err
	}

	return &rabbitMQService{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

func (r *rabbitMQService) Publish(ctx context.Context, queue string, message any) error {
	// Declare the queue
	_, err := r.channel.QueueDeclare(
		queue, // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to declare queue")
		return err
	}

	// Marshal the message
	body, err := json.Marshal(message)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to marshal message")
		return err
	}

	// Publish
	err = r.channel.PublishWithContext(ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)

	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to publish message")
		return err
	}

	return nil
}

func (r *rabbitMQService) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
	// Declare the queue
	_, err := r.channel.QueueDeclare(
		queue, // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to declare queue")
		return err
	}

	// Consume
	msgs, err := r.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to consume message")
		return err
	}

	go func() {
		for {
			select {
			case msgs, ok := <-msgs:
				if !ok {
					return
				}

				if err := handler(msgs.Body); err != nil {
					r.logger.Error().Err(err).Msg("Failed to handle message")
					msgs.Nack(false, true)
				} else {
					msgs.Ack(false)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (r *rabbitMQService) Close() error {
	// Close the channel
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			r.logger.Error().Err(err).Msg("Failed to close channel")
			return err
		}
	}

	// Close the connection
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.logger.Error().Err(err).Msg("Failed to close connection")
			return err
		}
	}

	return nil
}
