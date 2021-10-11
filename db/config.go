package db

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

var (
	dbUser string
	dbPass string
	dbHost string
	dbPort int
	dbName string
)

func LoadConfigFromEnv() {
	dbUser = os.Getenv("MONGODB_USER")
	dbPass = os.Getenv("MONGODB_PASS")
	dbHost = os.Getenv("MONGODB_HOST")
	dbName = os.Getenv("MONGODB_NAME")
	var err error
	dbPort, err = strconv.Atoi(os.Getenv("MONGODB_PORT"))
	if err != nil {
		log.Panicln(err)
	}
}

func LoadConfigFromFlags(flagSet *flag.FlagSet) {
	flag.StringVar(&dbUser, "mongodb_user", "ungame", "mongodb user")
	flag.StringVar(&dbPass, "mongodb_pass", "secret", "mongodb password")
	flag.StringVar(&dbHost, "mongodb_host", "localhost", "mongodb host")
	flag.IntVar(&dbPort, "mongodb_port", 27017, "mongodb port")
	flag.StringVar(&dbName, "mongodb_name", "workflows", "mongodb database name")
}

type Config interface {
	URL() string
}

type config struct {
	dbUser string
	dbPass string
	dbHost string
	dbPort int
	dbName string
}

func NewConfig() Config {
	return &config{
		dbUser: dbUser,
		dbPass: dbPass,
		dbHost: dbHost,
		dbPort: dbPort,
		dbName: dbName,
	}
}

func (cfg *config) URL() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?w=majority", cfg.dbUser, cfg.dbPass, cfg.dbHost, cfg.dbPort, cfg.dbName)
}
