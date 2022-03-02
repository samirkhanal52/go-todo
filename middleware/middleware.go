package middleware

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/samirkhanal52/go-todo/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbCollection *mongo.Collection

func init() {
	loadEnv()
	newDBInstace()
}

func loadEnv() {
	if dbURI, ok := os.LookupEnv("DB_URI"); !ok {
		log.Println("DB_URI Not Found!")
		os.Setenv("DB_URI", "mongodb://localhost:27017")
		os.Setenv("DB_NAME", "todo_list")
		os.Setenv("DB_COLLECTION_NAME", "To Do List")
		os.Setenv("HOST", ":8080")
		log.Println("DB_URI Defined..")
	} else {
		log.Println("DB_URI Found!..", dbURI)
	}
}

func newDBInstace() {
	//DB Connection String
	dbConnectionString := os.Getenv("DB_URI")

	//Database Name
	dbName := os.Getenv("DB_NAME")

	//DB Collection Name
	dbCollectionName := os.Getenv("DB_COLLECTION_NAME")

	//Set Client Options
	clientOptions := options.Client().ApplyURI(dbConnectionString)

	//Connect MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	checkErr(err)

	//Check the connection
	err = client.Ping(context.TODO(), nil)
	checkErr(err)

	dbCollection = client.Database(dbName).Collection(dbCollectionName)

	log.Println("Collection Instance Created....")
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
}

//Get All To do List GetTodo route
func HandleFetchToDo(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("X-Content-Type-Options", "nosniff")
	payload := getTodoList()
	res.WriteHeader(payload.ResponseCode)
	json.NewEncoder(res).Encode(map[string]interface{}{
		"message": payload.ResponseMessage,
		"status":  payload.ResponseCode,
		"data":    payload.ResponseData,
	})
}

//get to do data from DB
func getTodoList() models.JsonErrorModel {
	toDos := []models.ToDoModel{}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	cursor, err := dbCollection.Find(ctx, bson.M{})
	defer cursor.Close(context.Background())

	if err != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Couldn't Fetch Data\n", err),
			ResponseCode:    http.StatusProcessing,
		}
		return errRes
	}

	if err := cursor.All(ctx, &toDos); err != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Couldn't Fetch Data\n", err),
			ResponseCode:    http.StatusProcessing,
		}
		return errRes
	}

	toDoList := []models.Todo{}

	for _, t := range toDos {
		toDoList = append(toDoList, models.Todo{
			ID:          primitive.NewObjectID().Hex(),
			Title:       t.Title,
			Description: t.Description,
			CreatedAt:   t.CreatedAt,
			IsCompleted: t.IsCompleted,
		})
	}

	errRes := models.JsonErrorModel{
		ResponseMessage: "Data Fetched Successfully..",
		ResponseCode:    http.StatusOK,
		ResponseData:    toDoList,
	}

	return errRes
}

//Add To Do list route
func HandleAddToDo(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Method", "POST")
	res.Header().Set("X-Content-Type-Options", "nosniff")

	var todo models.Todo

	if err := json.NewDecoder(req.Body).Decode(&todo); err != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Couldn't Fetch Data\n", err),
			ResponseCode:    http.StatusProcessing,
		}

		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	if todo.Title == "" {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("The title is required"),
			ResponseCode:    http.StatusProcessing,
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	newToDo := models.ToDoModel{
		ID:          primitive.NewObjectID(),
		Title:       todo.Title,
		Description: todo.Description,
		IsCompleted: false,
		CreatedAt:   time.Now(),
		Remarks:     "",
	}

	payload := insertToDo(newToDo)
	json.NewEncoder(res).Encode(map[string]interface{}{
		"message": payload.ResponseMessage,
		"status":  payload.ResponseCode,
		"data":    payload.ResponseData,
	})
}

//insert new to do into db
func insertToDo(newToDo models.ToDoModel) models.JsonErrorModel {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if _, insertErr := dbCollection.InsertOne(ctx, &newToDo); insertErr != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Failed To Save Data", insertErr),
			ResponseCode:    http.StatusProcessing,
		}
		return errRes
	}

	errRes := models.JsonErrorModel{
		ResponseID:      newToDo.ID.Hex(),
		ResponseMessage: "Data Saved Successfully",
		ResponseCode:    http.StatusOK,
	}
	return errRes
}

//Todo List Delete route
func HandleDeleteToDo(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Method", "DELETE")
	res.Header().Set("X-Content-Type-Options", "nosniff")

	id, ok := req.URL.Query()["id"]

	if !ok || len(id[0]) < 1 || !primitive.IsValidObjectID(id[0]) {
		errRes := models.JsonErrorModel{
			ResponseMessage: "Invalid request ID",
			ResponseCode:    http.StatusBadRequest,
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	//encode to hex string
	param := hex.EncodeToString([]byte(id[0]))

	payload := deleteToDoList(param)
	json.NewEncoder(res).Encode(map[string]interface{}{
		"message": payload.ResponseMessage,
		"status":  payload.ResponseCode,
		"data":    payload.ResponseData,
	})
}

//delete Todo List item from db
func deleteToDoList(id string) models.JsonErrorModel {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	//hex string to primitive.objectId
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objectId}

	if _, err := dbCollection.DeleteOne(ctx, filter); err != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Failed to delete data", err),
			ResponseCode:    http.StatusBadRequest,
		}
		return errRes
	}

	errRes := models.JsonErrorModel{
		ResponseID:      id,
		ResponseMessage: "Todo List Deleted Successfully",
		ResponseCode:    http.StatusOK,
	}
	return errRes
}

//Todo List Update
func HandleUpdateToDo(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Method", "PUT")
	res.Header().Set("X-Content-Type-Options", "nosniff")

	id, ok := req.URL.Query()["id"]

	if !ok || len(id[0]) < 1 || !primitive.IsValidObjectID(id[0]) {
		errRes := models.JsonErrorModel{
			ResponseMessage: "Invalid request ID",
			ResponseCode:    http.StatusBadRequest,
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	var uTodo models.Todo

	if err := json.NewDecoder(req.Body).Decode(&uTodo); err != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Failed To Fetch Data", err),
			ResponseCode:    http.StatusProcessing,
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	if uTodo.Title == "" {
		errRes := models.JsonErrorModel{
			ResponseMessage: "Title Field is required",
			ResponseCode:    http.StatusProcessing,
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	if uTodo.Description == "" {
		errRes := models.JsonErrorModel{
			ResponseMessage: "Description Field is required",
			ResponseCode:    http.StatusProcessing,
		}
		json.NewEncoder(res).Encode(map[string]interface{}{
			"message": errRes.ResponseMessage,
			"status":  errRes.ResponseCode,
			"data":    errRes.ResponseData,
		})
		return
	}

	param := hex.EncodeToString([]byte(id[0]))
	payload := updateToList(param, uTodo)

	json.NewEncoder(res).Encode(map[string]interface{}{
		"message": payload.ResponseMessage,
		"status":  payload.ResponseCode,
		"data":    payload.ResponseID,
	})
}

//update to do list item
func updateToList(id string, uTodo models.Todo) models.JsonErrorModel {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objectId}

	if _, err := dbCollection.ReplaceOne(
		ctx,
		filter,
		bson.M{
			"title":       uTodo.Title,
			"iscompleted": uTodo.IsCompleted,
			"description": uTodo.Description,
			"updatedat":   uTodo.UpdatedAt,
		},
	); err != nil {
		errRes := models.JsonErrorModel{
			ResponseMessage: fmt.Sprint("Failed To Update To Do List", err),
			ResponseCode:    http.StatusProcessing,
		}
		return errRes
	}

	errRes := models.JsonErrorModel{
		ResponseID:      id,
		ResponseMessage: "Data Updated Successfully..",
		ResponseCode:    http.StatusOK,
	}
	return errRes
}
