package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	MONGO_PASSWORD := os.Getenv("MONGO_PASSWORD")
	MongoDB := fmt.Sprintf(`mongodb+srv://Everybody:%v@cluster0.wijun.mongodb.net/?retryWrites=true&w=majority`, MONGO_PASSWORD)
	fmt.Println(MongoDB)
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
		panic(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		panic(err.Error())
	}

	fmt.Println(client)

	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, collection_name string) *mongo.Collection {
	collection := client.Database("restaurant").Collection(collection_name)

	return collection
}
