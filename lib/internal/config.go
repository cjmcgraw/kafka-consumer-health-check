package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/thedevsaddam/gojsonq"
)

type Config struct {
	KafkaServers string
	GroupId      string
	Topics       []string
}

func LoadConfig(args ...string) *Config {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against consumer")
	jsonFilePath := parser.String("", "from-json-file", &argparse.Options{
		Default: "",
		Help:    "json file to use to load configurations. Has additional required arguments for jq keys",
	})

	fromEnvironment := parser.Flag("", "from-environmental-variables", &argparse.Options{
		Default: false,
		Help:    "load the configuration from environmental variables provided. Has additional requirements for keys",
	})

	parser.Parse(args)

	if len(*jsonFilePath) > 0 {
		log.Printf("loading config from json file=%s\n", *jsonFilePath)
		return loadFromFile(args, *jsonFilePath)
	}

	if *fromEnvironment {
		log.Println("loading config from environmental variables")
		return loadFromEnvironment(args)
	}

	log.Println("loading config from CLI arguments")
	return loadFromArgs(args)
}

func loadFromEnvironment(args []string) *Config {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against consumer")
	kafkaKey := parser.String("", "kafka-bootstrap-servers-from-env-key", &argparse.Options{
		Required: true,
		Help:     "key to load from the environmental variables",
	})
	groupIdKey := parser.String("", "kafka-consumer-group-id-from-env-key", &argparse.Options{
		Required: true,
		Help:     "key to load from the environmental variables",
	})
	topicsKey := parser.String("", "kafka-topics-from-env-key", &argparse.Options{
		Required: true,
		Help:     "key to load from the environmental variables",
	})
	parser.Parse(args)

	topicsCsv := os.Getenv(*topicsKey)
	topics := strings.Split(strings.TrimSpace(topicsCsv), ",")
	if len(topics) == 0 {
		msg := fmt.Sprintf("Failed to find valid envvar for --kafka-topics-from-env-key=%s", *topicsKey)
		log.Fatalln(parser.Usage(msg))
	}

	kafkaServers := os.Getenv(*kafkaKey)
	if len(kafkaServers) == 0 {
		msg := fmt.Sprintf("failed to find valid envvar for --kafka-bootstrap-servers-from-env-key=%s", *kafkaKey)
		log.Fatalln(parser.Usage(msg))
	}

	groupId := os.Getenv(*groupIdKey)
	if len(groupId) == 0 {
		msg := fmt.Sprintf("failed to find valid envvar for --kafka-consumer-group-id-from-env-key=%s", *groupIdKey)
		log.Fatalln(parser.Usage(msg))
	}

	return &Config{
		kafkaServers,
		groupId,
		topics,
	}
}

func loadFromFile(args []string, filepath string) *Config {
	jq := gojsonq.New().File(filepath)
	if jq.Error() != nil {
		log.Fatalf("error processing JQ file filepath=%s error=%s", filepath, jq.Error())
	}

	parser := argparse.NewParser("healthcheck", "runs healthcheck against")
	kafkaKey := parser.String("", "kafka-bootstrap-servers-from-json-file", &argparse.Options{
		Required: true,
		Help:     "jq accessible key format for the json file",
	})
	groupIdKey := parser.String("", "kafka-consumer-group-id-from-json-file", &argparse.Options{
		Required: true,
		Help:     "jq accessible key format for the json file",
	})
	topicKey := parser.String("", "kafka-topic-from-json-file", &argparse.Options{
		Required: true,
		Help:     "jq accessble key format for the json file",
	})
	parser.Parse(args)

	var topics []string
	jq.From(*topicKey).Where(*topicKey, "lengt", 0).Out(&topics)
	if jq.Error() != nil {
		log.Fatalf("failed to process topics from key file: err=%s\n", jq.Error())
	}
	if len(topics) <= 0 {
		msg := fmt.Sprintf("failed to process topic from key file. Invalid or missing key!")
		log.Fatalln(parser.Usage(msg))
	}

	kafkaServers := jq.From(*kafkaKey).String()
	if jq.Error() != nil {
		log.Printf("failed to process jq with error: %s", jq.Error())
		log.Fatalf("Failed to find key in json file via jq --kafka-bootstrap-servers-from-file=%s", *kafkaKey)
	}
	if len(kafkaServers) == 0 {
		log.Fatalf("Found empty/misisng key in json file using --kafka-bootstrap-servers-from-file=%s", *kafkaKey)
	}

	groupId := jq.From(*groupIdKey).String()
	if jq.Error() != nil {
		log.Printf("failed to process jq with error: %s", jq.Error())
		log.Fatalf("Failed to find key in json file via jq --kafka-consumer-group-id-from-json-file=%s", *groupIdKey)
	}
	if len(groupId) == 0 {
		log.Fatalf("Found empty/misisng key in json file using --kafka-consumer-group-id-from-file=%s", *groupIdKey)
	}

	return &Config{
		kafkaServers,
		groupId,
		topics,
	}
}

func loadFromArgs(args []string) *Config {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against the kafka consumer")

	KafkaServers := parser.String("", "kafka-bootstrap-servers", &argparse.Options{Required: true})
	GroupId := parser.String("", "kafka-consumer-group-id", &argparse.Options{Required: true})
	topicsCsv := parser.String("", "kafka-topics-csv", &argparse.Options{Required: true})
	parser.Parse(args)

	topics := strings.Split(strings.TrimSpace(*topicsCsv), ",")
	if len(topics) <= 0 {
		log.Fatalln(parser.Usage("invalid or missing parameter --kafka-topics-csv"))
	}

	if len(*KafkaServers) <= 0 {
		log.Fatalln(parser.Usage("missing required parameter --kafka-bootstrap-servers"))
	}

	if len(*GroupId) <= 0 {
		log.Fatalln(parser.Usage("missing required parameter --kafka-consumer-group-id"))
	}

	return &Config{
		KafkaServers: *KafkaServers,
		GroupId:      *GroupId,
		Topics:       topics,
	}
}
