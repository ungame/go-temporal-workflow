package api

import (
	"context"
	"flag"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"log"
	"time"
	"go-temporal-workflow/api/handlers"
	"go-temporal-workflow/db"
	"go-temporal-workflow/rmq"
	"go-temporal-workflow/services/subscriptions"
	"go-temporal-workflow/services/users"
	"go-temporal-workflow/store"
	"go-temporal-workflow/utils"
)

var port int

func init() {
	err := godotenv.Load(utils.GetEnvFilePath())
	if err != nil {
		log.Panicln(err)
	}
	flag.IntVar(&port, "api_port", 8080, "set api port")
	db.LoadConfigFromFlags(flag.CommandLine)
	rmq.LoadConfigFromFlags(flag.CommandLine)
	flag.Parse()
}

func Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbConn := db.New(ctx, db.NewConfig())
	err := dbConn.Ping(ctx)
	if err != nil {
		log.Panicln(err)
	}
	defer dbConn.Close(ctx)
	log.Println("mongodb connected!")

	rmqClient := rmq.NewClient(rmq.NewConfig())
	defer rmqClient.Close()
	log.Println("rabbitmq connected!")

	temporalClient, err := client.NewClient(client.Options{})
	if err != nil {
		log.Panicln(err)
	}
	defer temporalClient.Close()

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	usersStore := store.NewUsersStore(dbConn.DB())
	usersService := users.NewService(usersStore)

	subscriptionsStore := store.NewSubscriptionsStore(dbConn.DB())
	subscriptionsService := subscriptions.NewService(usersStore, subscriptionsStore, temporalClient)

	handlers.NewUsersHandlers(usersService, app)
	handlers.NewSubscriptionsHandler(rmqClient.AsPublisher(), subscriptionsService, app)

	err = app.Listen(":8080")
	if err != nil {
		log.Panicln(err)
	}
}
