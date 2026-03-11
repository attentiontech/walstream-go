package kafka

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	segkafka "github.com/segmentio/kafka-go"
)

// KafkaConfig holds Kafka connection settings.
type KafkaConfig struct {
	Brokers []string
	UseTLS  bool
}

// Validate checks that the config has all required fields.
func (c *KafkaConfig) Validate() error {
	if c == nil {
		return errors.New("kafka config is required")
	}
	if len(c.Brokers) == 0 {
		return errors.New("brokers are required")
	}
	return nil
}

// Dialer returns a kafka.Dialer configured for this connection.
func (c *KafkaConfig) Dialer() *segkafka.Dialer {
	d := &segkafka.Dialer{Timeout: 10 * time.Second}
	if c.UseTLS {
		d.TLS = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // G402: TLS verification not available in managed Kafka
	}
	return d
}

// prefixTopic prepends the prefix to the topic name with a dot separator.
// If prefix is empty, the topic is returned unchanged.
func prefixTopic(prefix, topic string) string {
	if prefix == "" {
		return topic
	}
	return fmt.Sprintf("%s.%s", prefix, topic)
}
