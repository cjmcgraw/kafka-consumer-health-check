package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/Shopify/sarama"
)

func getEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		errMsg := fmt.Sprintf("Missing required environmental variable: %s", key)
		panic(errMsg)
	}
	return val
}

func main() {
	kafkaBootstrapServerEnv := os.Getenv("KAFKA_BOOSTRAP_SERVER")
	kafkaTopicEnv := os.Getenv("KAFKA_TOPIC")
	kafkaConsumerGroupEnv := os.Getenv("KAFKA_CONSUMER_GROUP_ID")

	os.A

	parser := argparse.NewParser("kafka-consumer-health-check", "Checks the lag of the current event consumer and ensures it has not exceeded the given conditions")
	maxLag := *parser.Int("L", "maximum-allowed-lag",
		&argparse.Options{
			Required: true,
			Help: "the maximum number of events to lag behind before restarting, ignoring when the last event was processed"
	})

	allowedLag := *parser.Int("l", "acceptable-lag",
		&argparse.Options{
			Default: 0,
			Help: "the minimum amount of lag acceptable before checks are started",
	})

	maximumDelay := *parser.Int("D", "maximum-delay-seconds",
		&argparse.Options{
			Required: true,
			Help: "maximum number of seconds to have from last offset with lag before failling",
	})

	kafkaBootstrapServerArg := *parser.String("k", "kafka-boostrap-server",
		&argparse.Options{
			Required: len(kafkaBootstrapServerEnv) <= 0,
			Help: fmt.Sprintf("Kafka bootstrap server to use (default=%s)", kafkaBootstrapServerEnv),
	})

	kafkaTopicArg := *parser.String("t", "kafka-topic",
		&argparse.Options{
			Required: len(kafkaTopicEnv) <= 0,
			Help: fmt.Sprintf("Kafka topic to use (default=%s)", kafkaBootstrapServerEnv),
	})

	kafkaConsumerGroupArg := *parser.String("g", "kafka-consumer-group-id",
		&argparse.Options{
			Required: len(kafkaConsumerGroupEnv) <= 0,
			Help: fmt.Sprintf("Kafka consumer group id to use (default=%s)", kafkaConsumerGroupEnv),
	})

	err := parser.Parse(os.Args)
	if err != nil {
		panic(parser.Usage(err))
	}


	admin.GetMetadata()
	admin.DescribeConfigs()
}
