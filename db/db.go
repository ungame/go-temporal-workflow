package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

type Connection interface {
	DB() *mongo.Database
	Ping(ctx context.Context) error
	Close(ctx context.Context)
}

type connection struct {
	client *mongo.Client
}

func New(ctx context.Context, cfg Config) Connection {
	clientOptions := options.Client().ApplyURI(cfg.URL())
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Panicln(err)
	}
	return &connection{client: client}
}

func (c *connection) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

func (c *connection) Close(ctx context.Context) {
	err := c.client.Disconnect(ctx)
	if err != nil {
		log.Println("error on disconnect mongodb:", err)
	}
}

func (c *connection) DB() *mongo.Database {
	return c.client.Database(dbName)
}