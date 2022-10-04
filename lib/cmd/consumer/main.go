package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/confluentinc/confluent-kafka-go/kafka"
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
	parser := argparse.NewParser("kafka-consumer", "kafka consumer")
	delayTimeSeconds := parser.Int("d", "delay", &argparse.Options{Required: true, Help: "seconds to wait before processing each event"})
	kafkaBootstrapServer := getEnv("KAFKA_BOOTSTRAP_SERVER")
	kafkaGroupId := getEnv("KAFKA_GROUP_ID")
	kafkaTopic := getEnv("KAFKA_TOPIC")

	fmt.Printf("checking for exiting topic=%s\n", kafkaTopic)
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBootstrapServer,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results, err := admin.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             kafkaTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
		kafka.SetAdminOperationTimeout(time.Duration(3000)))
	if err != nil {
		panic(err)
	}
	fmt.Printf("created topics with response=%s\n", results)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBootstrapServer,
		"group.id":          kafkaGroupId,
		"auto.offset.reset": "earliest",
	})

	fmt.Printf("starting up consumer\n")
	c.Subscribe(kafkaTopic, nil)
	c.OffsetsForTimes()

	c.GetWatermarkOffsets()
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("inbound valid message. Sleeping for %d seconds...\n", *delayTimeSeconds)
			time.Sleep(time.Duration(*delayTimeSeconds))
			fmt.Printf("Processed message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			fmt.Printf("error processing message: %s\n", err)
		}
	}
	c.Close()
}
