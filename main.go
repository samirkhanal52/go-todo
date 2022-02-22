package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	hostName       string = "mongodb://localhost:27017"
	dbName         string = "todo_list"
	collectionName string = "To Do List"
	port           string = ":8080"
)

var tmpl *template.Template

type (
	toDoModel struct {
		ID          primitive.ObjectID `bson:"_id"`
		Title       string             `bson: "title"`
		Description string             `bson: "description"`
		IsCompleted bool               `bson: "is_completed"`
		CreatedAt   time.Time          `bson: "created_at"`
		UpdatedAt   time.Time          `bson: "updated_at"`
		Remarks     string             `bson: "remarks"`
	}

	todo struct {
		ID          primitive.ObjectID `json:"_id,omitempty"`
		Title       string             `json: "title"`
		Description string             `json: "description"`
		IsCompleted bool               `json: "is_completed"`
		CreatedAt   time.Time          `json: "created_at"`
		UpdatedAt   time.Time          `json: "updated_at"`
		Remarks     string             `json: "remarks"`
	}
)

func init() {
	// http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static/"))))

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := http.NewServeMux()
	tmpl = template.Must(template.ParseFiles("static/index.gohtml"))
	mux.HandleFunc("/", handleIndex)

	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	fmt.Println("Listening on port...", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}

	<-stopChan
	log.Println("Server Shutting Down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)

	defer cancel()
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
	case "POST":
	case "DELETE":
	default:
		http.Error(res, "Method Not Allowed", 405)
	}
}

func NewConnection() (context.Context, mongo.Collection) {
	clientOptions := options.Client().ApplyURI(hostName)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	checkErr(err)

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	coll := client.Database(dbName).Collection(collectionName)

	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		log.Fatal("ERROR: ", err)
	// 	}
	// }()

	return ctx, *coll
}

func checkErr(err error) {
	if err != nil {
		// fmt.Println("Mongo.Connet() ERROR: ", err)
		log.Fatal("ERROR: ", err)
	}
}

func main() {

	fetchDataByID()

}

func insertToDo() {
	ctx, col := NewConnection()

	toDo := toDoModel{
		ID:          primitive.NewObjectID(),
		Title:       "Attend Meeting",
		Description: "Attend Meeting At the conference my afternoon about the presentation",
		IsCompleted: false,
		CreatedAt:   time.Now(),
		Remarks:     "",
	}

	// fmt.Println("toDo Type:", reflect.TypeOf(toDo))

	if result, insertErr := col.InsertOne(ctx, toDo); insertErr != nil {
		log.Fatal("Insert Error:", insertErr)
		os.Exit(1)
	} else {
		newID := result.InsertedID
		fmt.Println("Data Inserted successfully, newID", newID)
	}
}

func fetchDataByID() {
	ctx, col := NewConnection()

	title := "Attend Meeting"

	var result bson.M

	if err := col.FindOne(ctx, bson.D{{"title", title}}).Decode(&result); err == mongo.ErrNoDocuments {
		fmt.Printf("No List Found")
		return
	} else if err != nil {
		log.Fatal(err)
		return
	}

	if jsonData, err := json.MarshalIndent(result, "", "	"); err != nil {
		log.Fatal(err)
		return
	} else {
		fmt.Printf("%s\n", jsonData)
	}
}
