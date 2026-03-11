package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/attentiontech/walstream-go/streaming"
	stypes "github.com/attentiontech/walstream-go/streaming/types"
	segkafka "github.com/segmentio/kafka-go"
)

// KafkaTopicSourceConfig configures a KafkaTopicSource.
type KafkaTopicSourceConfig struct {
	KafkaConfig *KafkaConfig
	TopicPrefix string // optional prefix prepended to topic names (e.g. "staging")
	Topics      []string
	Group       string
	Handler     streaming.ChangeHandler
	Logger      *slog.Logger
}

// KafkaTopicSource consumes walstream change events from Kafka topics.
// It is the consumer counterpart to the server-side sink: a single reader
// subscribed to all topics via the same consumer group.
type KafkaTopicSource struct {
	config KafkaTopicSourceConfig
	logger *slog.Logger
	reader *segkafka.Reader
}

// NewKafkaTopicSource creates a new KafkaTopicSource.
func NewKafkaTopicSource(config KafkaTopicSourceConfig) (*KafkaTopicSource, error) {
	if err := config.KafkaConfig.Validate(); err != nil {
		return nil, err
	}
	if len(config.Topics) == 0 {
		return nil, errors.New("at least one topic is required")
	}
	if config.Group == "" {
		return nil, errors.New("consumer group is required")
	}
	if config.Handler == nil {
		return nil, errors.New("handler is required")
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	logger = logger.With("component", "kafka_topic_source")

	dialer := config.KafkaConfig.Dialer()

	kafkaErrorLogger := segkafka.LoggerFunc(func(format string, args ...any) {
		logger.Error(fmt.Sprintf(format, args...))
	})

	topics := config.Topics
	if config.TopicPrefix != "" {
		topics = make([]string, len(config.Topics))
		for i, t := range config.Topics {
			topics[i] = prefixTopic(config.TopicPrefix, t)
		}
	}

	readerConfig := segkafka.ReaderConfig{
		Brokers:          config.KafkaConfig.Brokers,
		GroupTopics:      topics,
		GroupID:          config.Group,
		MaxWait:          1 * time.Second,
		MinBytes:         1,
		MaxBytes:         10e6, // 10MB
		RebalanceTimeout: 5 * time.Second,
		ErrorLogger:      kafkaErrorLogger,
		Dialer:           dialer,
	}

	return &KafkaTopicSource{
		config: config,
		logger: logger,
		reader: segkafka.NewReader(readerConfig),
	}, nil
}

// Consume starts consuming from all topics.
// It blocks until the context is cancelled.
func (s *KafkaTopicSource) Consume(ctx context.Context) error {
	s.logger.Info("starting consumer", "topics", s.config.Topics)
	defer s.logger.Info("consumer stopped")

	for {
		msg, err := s.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("failed to fetch message: %w", err)
		}

		var event stypes.ChangeEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			s.logger.Error("failed to unmarshal change event",
				"topic.name", msg.Topic,
				"offset", msg.Offset,
				"error", err)
			continue
		}

		if err := s.config.Handler.HandleChange(ctx, event); err != nil {
			return fmt.Errorf("failed to handle change event at offset %d on %s: %w", msg.Offset, msg.Topic, err)
		}

		if err := s.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("failed to commit offset %d: %w", msg.Offset, err)
		}
	}
}

// Close closes the Kafka reader.
func (s *KafkaTopicSource) Close() error {
	return s.reader.Close()
}
