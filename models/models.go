package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ToDoModel struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		Title       string             `bson:"title,omitempty"`
		Description string             `bson:"description"`
		IsCompleted bool               `bson:"is_completed,omitempty"`
		CreatedAt   time.Time          `bson:"created_at,omitempty"`
		UpdatedAt   time.Time          `bson:"updated_at"`
		Remarks     string             `bson:"remarks"`
	}

	Todo struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		IsCompleted bool      `json:"is_completed"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Remarks     string    `json:"remarks"`
	}

	JsonErrorModel struct {
		ResponseID      string `json:"response_id"`
		ResponseCode    int    `json:"status"`
		ResponseMessage string `json:"message"`
		ResponseData    []Todo `json:"data"`
	}
)
