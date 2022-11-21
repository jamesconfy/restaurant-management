package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {
	MONGO_PASSWORD := os.Getenv("MONGO_PASSWORD")
	MongoDB := fmt.Sprintf(`mongodb+srv://Everybody:%s@cluster0.wijun.mongodb.net/?retryWrites=true&w=majority`, MONGO_PASSWORD)

	fmt.Print(MongoDB)
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
		panic(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		panic(err.Error())
	}

	fmt.Println("Connected to MongoDB")

	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, collection_name string) *mongo.Collection {
	collection := client.Database("restaurant").Collection(collection_name)

	return collection
}
