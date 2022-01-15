package store

import (
	"context"
	"log"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/utils"
)

var client *mongo.Client
var cancel context.CancelFunc

func init() {
	log.Println("Running version:", consts.Version, "builded:", consts.Build, "commit:", consts.Commit)
	url := utils.GetEnv("MONGO_URL", "mongodb://admin:password@localhost:27017")
	log.Println("mongodb connection url", url)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	cancel = cancelFunc
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	client = c
	if err != nil {
		log.Fatal("connecting to a database:", url, "failed", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("pinging a database: ", url, "failed", err)
	}
	log.Println("pinged a database succesfully")
}

func GetClient() *mongo.Client {
	return client
}

func Close() {
	cancel()
	client.Disconnect(context.Background())
}

// SetupTest clears database
func SetupTest(t *testing.T) func() {
	// Create new session
	sess, err := client.StartSession()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()

	// This will be returned to be called on upper level (in tests)
	teardown := func() {
		sess.EndSession(ctx)
	}

	// With newly created session...
	err = client.UseSession(ctx, func(sessCtx mongo.SessionContext) error {
		// Delete documents from all collections
		for _, coll := range consts.AllCollections {
			client.Database(consts.DatabaseName).Collection(coll).DeleteMany(context.TODO(), bson.M{})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return teardown
}
