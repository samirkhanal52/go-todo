package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"
	"string"
	

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoField struct {
	ToDoName         string //json: "Field String"
	TodoId           int    //json: "Field int"
	TodDoDescription string //json: "Field String"
	ToDoStatus       bool   //json: "Field Bool"
}

func main() {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// fmt.Println("ClientOption TYPE:", reflect.TypeOf(clientOptions))

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println("Mongo.Connet() ERROR: ", err)
		os.Exit(1)
	}
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	col := client.Database("todo_list").Collection("To Do List")
	fmt.Println("Collection Type: ", reflect.TypeOf(col))

	oneDoc := MongoField{
		ToDoName:         "Attend Meeting",
		TodoId:           1,
		TodDoDescription: "Attend Meeting At the conference my afternoon about the presentation",
		ToDoStatus:       false,
	}

	fmt.Println("oneDoc Type:", reflect.TypeOf(oneDoc))

	result, insertErr := col.InsertOne(ctx, oneDoc)
	if insertErr != nil {
		fmt.Println("InsertOne Error:", insertErr)
		os.Exit(1)
	} else {
		fmt.Println("InsertOne() result type: ", reflect.TypeOf(result))
		fmt.Println("InsertOne() API result type: ", result)

		newID := result.InsertedID
		fmt.Println("InsertedOne(), newID", newID)
		fmt.Println("InsertedOne(), newID Type", reflect.TypeOf(newID))
	}

}
