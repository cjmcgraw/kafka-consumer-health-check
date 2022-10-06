package internal

import (
	"github.com/akamensky/argparse"
	"github.com/thedevsaddam/gojsonq"
	"log"
	"os"
)

type Config struct {
	KafkaServers string
	GroupId      string
}

func LoadConfig(arguments []string) *Config {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against consumer")
	jsonFilePath := *parser.String("", "--from-json-file", &argparse.Options{
		Default: "",
		Help:    "json file to use to load configurations. Has additional required arguments for jq keys",
	})

	fromEnvironment := *parser.Flag("", "--from-environmental-variables", &argparse.Options{
		Default: false,
		Help:    "load the configuration from environmental variables provided. Has additional requirements for keys",
	})

	parser.Parse(arguments)

	if len(jsonFilePath) >= 0 {
		return loadFromFile(arguments, jsonFilePath)
	}

	if fromEnvironment {
		return loadFromEnvironment(arguments)
	}

	return loadFromArgs(arguments)
}

func loadFromEnvironment(args []string) *Config {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against consumer")
	kafkaKey := *parser.String("", "--kafka-bootstrap-servers-from-env-key", &argparse.Options{
		Required: true,
		Help:     "key to load from the environmental variables",
	})
	groupIdKey := *parser.String("", "--kafka-consumer-group-id-from-env-key", &argparse.Options{
		Required: true,
		Help:     "key to load from the environmental variables",
	})
	parser.Parse(args)

	kafkaServers := os.Getenv(kafkaKey)
	if len(kafkaServers) == 0 {
		log.Fatalf("failed to find envvar for key --kafka-bootstrap-servers-from-env-key=%s", kafkaKey)
	}

	groupId := os.Getenv(groupIdKey)
	if len(groupId) == 0 {
		log.Fatalf("Failed to find envvar for --kafka-consumer-group-id-from-env-key=%s", groupIdKey)
	}

	return &Config{
		kafkaServers,
		groupId,
	}
}

func loadFromFile(args []string, filepath string) *Config {
	jq := gojsonq.New().File(filepath)
	if jq.Error() != nil {
		panic(jq.Errors())
	}

	parser := argparse.NewParser("healthcheck", "runs healthcheck against")
	kafkaKey := *parser.String("", "--kafka-bootstrap-servers-from-json-file", &argparse.Options{
		Required: true,
		Help:     "jq accessible key format for the json file",
	})
	groupIdKey := *parser.String("", "--kafka-consumer-group-id-from-json-file", &argparse.Options{
		Required: true,
		Help:     "jq accessible key format for the json file",
	})
	parser.Parse(args)

	kafkaServers := jq.From(kafkaKey).String()
	if jq.Error() != nil {
		log.Printf("failed to process jq with error: %s", jq.Error())
		log.Fatalf("Failed to find key in json file via jq --kafka-bootstrap-servers-from-file=%s", kafkaKey)
	}
	if len(kafkaServers) == 0 {
		log.Fatalf("Found empty/misisng key in json file using --kafka-bootstrap-servers-from-file=%s", kafkaKey)
	}

	groupId := jq.From(groupIdKey).String()
	if jq.Error() != nil {
		log.Printf("failed to process jq with error: %s", jq.Error())
		log.Fatalf("Failed to find key in json file via jq --kafka-consumer-group-id-from-json-file=%s", groupIdKey)
	}
	if len(groupId) == 0 {
		log.Fatalf("Found empty/misisng key in json file using --kafka-consumer-group-id-from-file=%s", groupIdKey)
	}

	return &Config{
		kafkaServers,
		groupId,
	}
}

func loadFromArgs(args []string) *Config {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against the kafka consumer")
	config := &Config{
		KafkaServers: *parser.String("", "kafka-bootstrap-servers", &argparse.Options{Required: true}),
		GroupId:      *parser.String("", "kafka-consumer-group-id", &argparse.Options{Required: true}),
	}
	parser.Parse(args)

	if len(config.KafkaServers) <= 0 {
		log.Fatalln("missing required parameter --kafka-bootstrap-servers")
	}

	if len(config.GroupId) <= 0 {
		log.Fatalln("missing required parameter --kafka-consumer-group-id")
	}

	return config
}
