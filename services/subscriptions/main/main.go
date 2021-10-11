package main

import (
	"context"
	"flag"
	"github.com/streadway/amqp"
	"go.temporal.io/sdk/client"
	"log"
	"time"
	"go-temporal-workflow/db"
	"go-temporal-workflow/rmq"
	"go-temporal-workflow/services/subscriptions"
	"go-temporal-workflow/store"
)

func init() {
	db.LoadConfigFromFlags(flag.CommandLine)
	rmq.LoadConfigFromFlags(flag.CommandLine)
	flag.Parse()
}

func main() {
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

	err = rmqClient.ExchangeDeclare(&rmq.ExchangeOptions{
		Name:    "subscription",
		Kind:    amqp.ExchangeDirect,
		Durable: true,
	})
	if err != nil {
		log.Panicln(err)
	}

	err = rmqClient.QueueDeclare(&rmq.QueueOptions{
		Name:    "subscriptions",
		Durable: true,
		BindOptions: &rmq.QueueBindOptions{
			ExchangeName: "subscription",
		},
	})

	temporalClient, err := client.NewClient(client.Options{})
	if err != nil {
		log.Panicln(err)
	}
	defer temporalClient.Close()

	consumer := rmqClient.AsConsumer()

	usersStore := store.NewUsersStore(dbConn.DB())
	subscriptionsStore := store.NewSubscriptionsStore(dbConn.DB())
	subscriptionsService := subscriptions.NewService(usersStore, subscriptionsStore, temporalClient)
	subscriptions.NewHandler(subscriptionsService, consumer)

	go subscriptions.NewWorker(temporalClient, subscriptionsService)

	err = consumer.Listen(&rmq.ConsumerOptions{
		QueueName: "subscriptions",
	})
	if err != nil {
		log.Panicln(err)
	}
}
