package main

import (
	"github.com/akamensky/argparse"
	"github.com/carlm/kafka-consumer-health-check/internal"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"os"
	"time"
)

func main() {
	config := *internal.LoadConfig(os.Args)
	log.Printf("using config=%s\n", config)
	parser := argparse.NewParser("healthcheck", "allows healthchecking through consumer")
	maxDelayMs := int64(1000 * *parser.Int("", "--maximum-delay-seconds", &argparse.Options{
		Required: true,
		Help:     "maximum number of seconds to wait for an offset to be processed with lag before its considered failed",
	}))
	maxLag := *parser.Int("", "--maximum-lag-events", &argparse.Options{
		Default: 1_000,
		Help:    "maximum valid lag, exceeding this value in lag will be considered a failure",
	})
	parser.Parse(os.Args)

	if maxDelayMs <= 10_000 {
		log.Fatalln("Must provide a max delay of 10seconds or more! invalid --maximum-delay-seconds")
	}

	if maxLag <= 1 {
		log.Fatalln("Must provide a maxLag of more than 1! invalid --maximum-lag-events")
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.KafkaServers,
		"group.id":          config.GroupId,
	})
	if err != nil {
		log.Fatalf("failed during consumer creation err=%s\n", err)
	}
	log.Println("established consumer to kafka")

	partitions, err := consumer.Assignment()
	if err != nil {
		log.Fatalf("failed during partition selection for consumer err=%s\n", err)
	}
	log.Printf("found %d partitions=%s\n", len(partitions), partitions)
	log.Printf("querying for all topics that have committed an event in the last %d seconds\n", maxDelayMs)

	maximumLookbackTime := time.Now().UnixMilli() - maxDelayMs
	var offsetQueryPartitions []kafka.TopicPartition
	for _, partition := range partitions {
		queryPartition := kafka.TopicPartition{
			Topic:     partition.Topic,
			Partition: partition.Partition,
			Offset:    kafka.Offset(maximumLookbackTime),
		}
		offsetQueryPartitions = append(offsetQueryPartitions, queryPartition)
	}
	offsetsFromPartitionsByTime, err := consumer.OffsetsForTimes(offsetQueryPartitions, 10_000)
	if err != nil {
		log.Fatalf("failed to retrieve partitions lookback for unknown reason err=%s\n", err)
	}

	var hasWritten []bool
	for _, partitionWithTime := range offsetsFromPartitionsByTime {
		if partitionWithTime.Offset > 0 {
			hasWritten = append(hasWritten, true)
			log.Printf("found write in time period parition=%s", partitionWithTime)
		} else {
			log.Printf("no write in time period partition=%s", partitionWithTime)
		}
	}

	var currentOffsets []int64
	var highWatermarks []int64

	log.Println("checking partition offsets and lag...")
	for _, partition := range partitions {
		currentOffset := int64(partition.Offset)
		_, highWatermark, watermarkError := consumer.QueryWatermarkOffsets(
			*partition.Topic,
			partition.Partition,
			10_000,
		)

		if watermarkError != nil {
			log.Printf("failed to get high offset for partition=%s\n", partition)
		} else {
			log.Printf("found current=%d high=%d partition=%s", currentOffset, highWatermark, partition)
			currentOffsets = append(currentOffsets, currentOffset)
			highWatermarks = append(highWatermarks, highWatermark)
		}
	}
}
