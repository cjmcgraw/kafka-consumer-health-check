package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/carlm/kafka-consumer-health-check/internal"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	config := internal.LoadConfig(
		os.Args[0],
		"--from-environmental-variables",
		"--kafka-bootstrap-servers-from-env-key", "KAFKA_BOOTSTRAP_SERVER",
		"--kafka-consumer-group-id-from-env-key", "KAFKA_GROUP_ID",
		"--kafka-topics-from-env-key", "KAFKA_TOPIC",
	)
	log.Printf("loaded config=%+v\n", config)

	parser := argparse.NewParser("kafka-consumer", "kafka consumer")
	delayTimeSeconds := parser.Int("d", "delay", &argparse.Options{Required: true, Help: "seconds to wait before processing each event"})
	parser.Parse(os.Args)

	log.Printf("checking for exiting topic=%s\n", config.Topics[0])
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": config.KafkaServers,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results, err := admin.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             config.Topics[0],
			NumPartitions:     1,
			ReplicationFactor: 1,
		}},
		kafka.SetAdminOperationTimeout(time.Duration(3000)))
	if err != nil {
		panic(err)
	}
	log.Printf("created topics with response=%s\n", results)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.KafkaServers,
		"group.id":          config.GroupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("starting up consumer\n")
	err = c.Subscribe(config.Topics[0], nil)
	if err != nil {
		panic(err)
	}
	log.Printf("retrieving partition assignments")
	partitions, err := c.Assignment()
	if err !
	log.Printf("partitions=%v", partitions)
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			log.Printf("inbound valid message. Sleeping for %d seconds...\n", *delayTimeSeconds)
			time.Sleep(time.Duration(*delayTimeSeconds))
			log.Printf("Processed message on partition=%s: value=%x\n", msg.TopicPartition, string(msg.Value))
		} else {
			log.Printf("error processing message: %s\n", err)
		}
	}
}
