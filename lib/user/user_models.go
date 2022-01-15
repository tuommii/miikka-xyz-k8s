package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CollectionName = "users"

type User struct {
	ID          primitive.ObjectID `bson:"_id"`
	FirstName   string             `bson:"first_name"`
	LastName    string             `bson:"last_name"`
	FullName    string             `bson:"full_name"`
	Username    string             `bson:"username"`
	Age         int                `bson:"age"`
	DateOfBirth time.Time          `bson:"date_of_birth"`
	Avatar      string             `bson:"avatar,omitempty"`
	IsActive    bool               `bson:"is_active"`
}

type UserInput struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
}
