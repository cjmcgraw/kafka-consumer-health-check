package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/carlm/kafka-consumer-health-check/internal"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

func getUUID() []byte {
	val, err := uuid.New().MarshalBinary()
	if err != nil {
		log.Fatalf("failed to generate UUID for unknown reason! err=%s\n", err)
	}
	return val
}

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
	events := parser.Int("n", "events", &argparse.Options{Required: true, Help: "number of events to send"})
	parser.Parse(os.Args)

	if *events <= 0 {
		log.Fatalln("must provide at least 1 event with --events ")
	}

	log.Printf("checking for exiting topic=%s\n", config.Topics[0])
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": config.KafkaServers,
	})
	defer admin.Close()
	if err != nil {
		log.Fatalf("failed to create admin client! err=%s\n", err)
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
		log.Fatalf("failed to create topics for unknown reason, err=%s\n", err)
	}
	log.Printf("created topics with response=%s\n", results)

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.KafkaServers,
	})
	defer p.Close()
	if err != nil {
		log.Fatalf("failed to create producer! err=%s\n", err)
	}

	log.Printf("sending %d events\n", *events)
	for i := 0; i < *events; i++ {
		value := getUUID()
		log.Printf("sending event value=%x\n", value)
		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &config.Topics[0],
				Partition: kafka.PartitionAny,
			},
			Value: value,
		}, nil)
		log.Println("finished sending event")
	}
	log.Println("finished sending all events")
}
