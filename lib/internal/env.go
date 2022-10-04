package internal

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
	"reflect"
)

type Config struct {
	kafkaServers string
	groupId      string
	topic        string
}

type ArgsConfig struct {
	config     *Config
	configFile string
}

func LoadConfig(arguments []string) *Config {
	var configs []Config

	envConfig := *loadFromEnvironment()
	configs = append(configs, envConfig)

	args := *loadFromArgs(arguments)
	configs = append(configs, *args.config)
	if len(args.configFile) > 0 {
		fmt.Printf("found config file at %s", args.configFile)
		fileConfig := *loadFromFile(args.configFile)
		configs = append(configs, fileConfig)
	}

	getConfigValue := func(s string) string {
		for _, config := range configs {
			v := reflect.ValueOf(config).FieldByName(s)
			if v.IsValid() && v.Len() > 0 {
				return v.String()
			}
		}
		panic(fmt.Errorf("failed to find key %s in configuration", s))
	}

	return &Config{
		kafkaServers: getConfigValue("kafkaServers"),
		groupId:      getConfigValue("groupId"),
		topic:        getConfigValue("topic"),
	}
}

func loadFromEnvironment() *Config {
	return &Config{
		kafkaServers: os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		groupId:      os.Getenv("KAFKA_CONSUMER_GROUP_ID"),
		topic:        os.Getenv("KAFKA_TOPIC"),
	}
}

func loadFromFile(filepath string) *Config {
	return &Config{}
}

func loadFromArgs(args []string) *ArgsConfig {
	parser := argparse.NewParser("healthcheck", "runs healthcheck against the kafka consumer")
	argsConfig := &ArgsConfig{
		config: &Config{
			kafkaServers: *parser.String("s", "kafka-bootstrap-servers", &argparse.Options{Default: ""}),
			groupId:      *parser.String("g", "kafka-consumer-group-id", &argparse.Options{Default: ""}),
			topic:        *parser.String("t", "kafka-topic", &argparse.Options{Default: ""}),
		},
		configFile: *parser.String("f", "config-file", &argparse.Options{Default: ""}),
	}
	parser.Parse(args)
	return argsConfig
}
