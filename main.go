package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	port           string = ":8000"
)

var dbClient context.Context
var tmpl *template.Template

type (
	toDoModel struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		Title       string             `bson:"title,omitempty"`
		Description string             `bson:"description"`
		IsCompleted bool               `bson:"is_completed,omitempty"`
		CreatedAt   time.Time          `bson:"created_at,omitempty"`
		UpdatedAt   time.Time          `bson:"updated_at"`
		Remarks     string             `bson:"remarks"`
	}

	todo struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		IsCompleted bool      `json:"is_completed"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Remarks     string    `json:"remarks"`
	}

	jsonError struct {
		responseID      string `json:"response_id"`
		responseMessage string `json:"response_msg"`
		responseCode    int32  `json:"response_code"`
		responseErr     error  `json:"response_err"`
	}
)

func init() {
	clientOptions := options.Client().ApplyURI(hostName)
	_, err := mongo.Connect(context.TODO(), clientOptions)
	checkErr(err)
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	reqUrl := "." + req.URL.Path
	if reqUrl == "./" {
		log.Println(req.Method)
		reqUrl = "./static/index.html"

		http.ServeFile(res, req, reqUrl)

	} else if reqUrl == "./todo" {
		switch req.Method {
		case "GET":
			handleFetchToDo(res, req)
		case "POST":
			handleAddToDo(res, req)
		case "DELETE":
			handleDeleteToDo(res, req)
		case "PUT":
			handleUpdateToDo(res, req)
		default:
			http.NotFound(res, req)
		}
	} else {
		http.NotFound(res, req)
	}
}

func responseError(res http.ResponseWriter, data interface{}, responseCode int) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("X-Content-Type-Options", "nosniff")
	res.WriteHeader(responseCode)
	json.NewEncoder(res).Encode(data)
}

func handleFetchToDo(res http.ResponseWriter, req *http.Request) {
	toDos := []toDoModel{}

	dbCollection := NewConnection()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	cursor, err := dbCollection.Find(ctx, bson.M{})
	defer cursor.Close(ctx)

	if err != nil {
		errRes := jsonError{
			responseMessage: "Couldn't Fetch Data",
			responseCode:    http.StatusProcessing,
			responseErr:     err,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	if err := cursor.All(ctx, &toDos); err != nil {
		errRes := jsonError{
			responseMessage: "Couldn't Fetch Data",
			responseCode:    http.StatusProcessing,
			responseErr:     err,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	toDoList := []todo{}

	for _, t := range toDos {
		toDoList = append(toDoList, todo{
			ID:          primitive.NewObjectID().Hex(),
			Title:       t.Title,
			Description: t.Description,
			CreatedAt:   t.CreatedAt,
			IsCompleted: t.IsCompleted,
		})
	}

	responseError(res, toDoList, http.StatusOK)
}

func handleAddToDo(res http.ResponseWriter, req *http.Request) {
	var t todo

	dbCollection := NewConnection()
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := json.NewDecoder(req.Body).Decode(&t); err != nil {
		errRes := jsonError{
			responseMessage: "Couldn't Fetch Data",
			responseCode:    http.StatusProcessing,
			responseErr:     err,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	if t.Title == "" {
		errRes := jsonError{
			responseMessage: "The title is required",
			responseCode:    http.StatusProcessing,
			responseErr:     nil,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	newToDo := toDoModel{
		ID:          primitive.NewObjectID(),
		Title:       t.Title,
		IsCompleted: false,
		CreatedAt:   time.Now(),
		Remarks:     "",
	}

	if _, insertErr := dbCollection.InsertOne(ctx, &newToDo); insertErr != nil {
		errRes := jsonError{
			responseMessage: "Failed To Save Data",
			responseCode:    http.StatusProcessing,
			responseErr:     insertErr,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	errRes := jsonError{
		responseID:      newToDo.ID.Hex(),
		responseMessage: "Data Saved Successfully",
		responseCode:    http.StatusOK,
	}
	responseError(res, errRes, http.StatusOK)
}

func handleDeleteToDo(res http.ResponseWriter, req *http.Request) {
	id, ok := req.URL.Query()["ID"]

	if !ok || len(id[0]) < 1 || !primitive.IsValidObjectID(id[0]) {
		errRes := jsonError{
			responseMessage: "Invalid request ID",
			responseCode:    http.StatusBadRequest,
		}
		responseError(res, errRes, http.StatusBadRequest)
		return
	}

	dbCollection := NewConnection()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	objectId, _ := primitive.ObjectIDFromHex(hex.EncodeToString([]byte(id[0])))

	if _, err := dbCollection.DeleteOne(ctx, bson.M{"_id": objectId}); err != nil {
		errRes := jsonError{
			responseMessage: "Failed to delete data",
			responseCode:    http.StatusBadRequest,
			responseErr:     err,
		}
		responseError(res, errRes, http.StatusBadRequest)
		return
	}

	errRes := jsonError{
		responseID:      id[0],
		responseMessage: "Todo List Deleted Successfully",
		responseCode:    http.StatusOK,
	}
	responseError(res, errRes, http.StatusOK)
}

func handleUpdateToDo(res http.ResponseWriter, req *http.Request) {
	id, ok := req.URL.Query()["ID"]

	if !ok || len(id[0]) < 1 || !primitive.IsValidObjectID(id[0]) {
		errRes := jsonError{
			responseMessage: "Invalid request ID",
			responseCode:    http.StatusBadRequest,
		}
		responseError(res, errRes, http.StatusBadRequest)
		return
	}

	var uTodo todo

	if err := json.NewDecoder(req.Body).Decode(&uTodo); err != nil {
		errRes := jsonError{
			responseMessage: "Failed To Fetch Data",
			responseCode:    http.StatusProcessing,
			responseErr:     err,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	if uTodo.Title == "" {
		errRes := jsonError{
			responseMessage: "Title Field is required",
			responseCode:    http.StatusProcessing,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}

	dbCollection := NewConnection()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	objectId, _ := primitive.ObjectIDFromHex(hex.EncodeToString([]byte(id[0])))

	if _, err := dbCollection.ReplaceOne(
		ctx,
		bson.M{"_id": objectId},
		bson.M{
			"title":       uTodo.Title,
			"iscompleted": uTodo.IsCompleted,
			"description": uTodo.Description,
			"updatedat":   uTodo.UpdatedAt,
			"remarks":     uTodo.Remarks,
		},
	); err != nil {
		errRes := jsonError{
			responseMessage: "Failed To Update To Do List",
			responseCode:    http.StatusProcessing,
			responseErr:     err,
		}
		responseError(res, errRes, http.StatusProcessing)
		return
	}
}

func NewConnection() mongo.Collection {
	clientOptions := options.Client().ApplyURI(hostName)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	checkErr(err)

	coll := client.Database(dbName).Collection(collectionName)

	return *coll
}

func checkErr(err error) {
	if err != nil {
		// fmt.Println("Mongo.Connet() ERROR: ", err)
		log.Fatal("ERROR: ", err)
	}
}

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/", handleIndex)
	// http.HandleFunc("/todo", todoHandlers)
	// http.HandleFunc("/", handleFetchToDo)
	// srvMux.HandleFunc("/", handleAddToDo)
	// srvMux.HandleFunc("/{id}", handleDeleteToDo)
	// srvMux.HandleFunc("/{id}", handleUpdateToDo)
	// srvMux.HandleFunc("/getCompletedTask", handleCompletedToDo)

	srv := &http.Server{
		Addr:         port,
		Handler:      srvMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go func() {
		fmt.Println("Listening on port...", port)

		if err := http.ListenAndServe(port, srvMux); err != nil {
			log.Fatal(err)
		}
	}()

	<-stopChan
	log.Println("Server Shutting Down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	fmt.Println("Server Shut Down")
}

func insertToDo() {
	col := NewConnection()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		log.Fatal("ERROR: ", err)
	// 	}
	// }()

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
	col := NewConnection()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		log.Fatal("ERROR: ", err)
	// 	}
	// }()

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
