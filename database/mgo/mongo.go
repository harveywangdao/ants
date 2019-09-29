package mgo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/harveywangdao/ants/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MongoTimeout = 10 * time.Second
)

func NewMgoClient(addr, user, pw string) (*mongo.Client, error) {
	// mongodb://foo:bar@localhost:27017
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + user + ":" + pw + "@" + addr))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), MongoTimeout)
	//defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	//ctx, _ := context.WithTimeout(context.Background(), MongoTimeout)
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+addr))

	//ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	//err = client.Ping(ctx, readpref.Primary())

	return client, nil
}

func mgoTest(client *mongo.Client) {
	collection := client.Database("dbName").Collection("numbers")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	id := res.InsertedID
	fmt.Println(id)

	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("result:", result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	var result struct {
		Value float64
	}
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = collection.FindOne(ctx, bson.M{"name": "pi"}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("result:", result)
}
