package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
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

func getUUID() []byte {
	val, err := uuid.New().MarshalBinary()
	if err != nil {
		panic(err)
	}
	return val
}

func main() {
	parser := argparse.NewParser("kafka-consumer", "kafka consumer")
	events := parser.Int("n", "events", &argparse.Options{Required: true, Help: "number of events to send"})
	err := parser.Parse(os.Args)
	if err != nil {
		panic(parser.Usage(err))
	}
	kafkaBootstrapServer := getEnv("KAFKA_BOOTSTRAP_SERVER")
	kafkaTopic := getEnv("KAFKA_TOPIC")

	fmt.Printf("%d\n", *events)
	if *events <= 0 {
		panic("must provide atleast 1 event with --events ")
	}

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

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBootstrapServer,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("sending %d events\n", *events)
	for i := 0; i < *events; i++ {
		value := getUUID()
		fmt.Printf("sending event %s\n", value)
		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &kafkaTopic, Partition: kafka.PartitionAny},
			Value:          value,
		}, nil)
		fmt.Println("finished sending event")
	}
	fmt.Println("finished sending all events")
	p.Close()
}
