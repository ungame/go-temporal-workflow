package rmq

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

var (
	rmqUser string
	rmqPass string
	rmqHost string
	rmqPort int
)

func LoadConfigFromEnv() {
	rmqUser = os.Getenv("RABBITMQ_USER")
	rmqPass = os.Getenv("RABBITMQ_PASS")
	rmqHost = os.Getenv("RABBITMQ_HOST")
	var err error
	rmqPort, err = strconv.Atoi(os.Getenv("RABBITMQ_PORT"))
	if err != nil {
		log.Panicln(err)
	}
}

func LoadConfigFromFlags(flagSet *flag.FlagSet) {
	flag.StringVar(&rmqUser, "rmq_user", "guest", "rabbitmq user")
	flag.StringVar(&rmqPass, "rmq_pass", "guest", "rabbitmq password")
	flag.StringVar(&rmqHost, "rmq_host", "localhost", "rabbitmq host")
	flag.IntVar(&rmqPort, "rmq_port", 5672, "rabbitmq port")
}

type Config interface {
	URL() string
}

type config struct {
	rmqUser string
	rmqPass string
	rmqHost string
	rmqPort int
}

func NewConfig() Config {
	return &config{
		rmqUser: rmqUser,
		rmqPass: rmqPass,
		rmqHost: rmqHost,
		rmqPort: rmqPort,
	}
}

func (cfg *config) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.rmqUser, cfg.rmqPass, cfg.rmqHost, cfg.rmqPort)
}
