package user

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/store"
)

func StoreCreateUser(user *UserInput) (primitive.ObjectID, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionUsers)
	res, err := coll.InsertOne(context.TODO(), user)
	if err != nil {
		log.Println(err)
		return primitive.NilObjectID, err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Println("should never happen: ", consts.ErrInvalidID)
		return primitive.NilObjectID, errors.New(consts.ErrInvalidID)
	}
	return id, nil
}

func StoreGetUserByUsername(username string) (*User, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionUsers)
	res := coll.FindOne(context.TODO(), bson.M{"username": username})
	err := res.Err()
	// Not found "error". This needs to be handled seperatly
	if err != nil && err.Error() == mongo.ErrNoDocuments.Error() {
		return nil, nil
	}
	// "Real" error
	if err != nil {
		log.Println(err)
		return nil, err
	}

	user := User{}
	err = res.Decode(&user)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &user, nil
}

func StoreGetUsersByUsername(usernames []string) ([]User, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionUsers)
	users := make([]User, 0)

	cursor, err := coll.Find(context.TODO(), bson.M{"username": bson.M{"$in": usernames}})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = cursor.All(context.TODO(), &users)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return users, nil
}

func StoreGetUsersByID(ids []primitive.ObjectID) ([]User, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionUsers)
	users := make([]User, 0)

	cursor, err := coll.Find(context.TODO(), bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = cursor.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = cursor.Decode(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
