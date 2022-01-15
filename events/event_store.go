package events

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/store"
)

func StoreCreateEvent(ctx context.Context, event *Event) (primitive.ObjectID, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionEvents)
	event.ID = primitive.NewObjectID()
	res, err := coll.InsertOne(ctx, event)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New(consts.ErrInvalidID)
	}
	return id, nil
}

func StoreGetEventByID(ctx context.Context, id primitive.ObjectID) (*Event, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionEvents)

	event := &Event{}
	res := coll.FindOne(ctx, bson.M{"_id": id})
	// Not found "error"
	/*if res.Err() != nil && res.Err().Error() == mongo.ErrNoDocuments.Error() {
		return nil, nil
	}*/
	// "Real" error
	if res.Err() != nil {
		return nil, res.Err()
	}

	return event, res.Decode(&event)
}
