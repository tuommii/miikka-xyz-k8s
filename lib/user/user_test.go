package user

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/events"
	"miikka.xyz/devops-app/store"
)

// Test real endpoint and expect events to be created also
func TestHandleCreateUser(t *testing.T) {
	teardown := store.SetupTest(t)
	defer teardown()

	rabbitConn, rabbitCh := events.CreateQueue(consts.QueueEventsName)
	defer rabbitCh.Close()
	defer rabbitConn.Close()

	tt := []struct {
		username string
		status   int
	}{
		{"tuommii", 200},
		{"tuommii", 400},
		{"", 400},
		{"\t\t\t\t\r\r\r\"\"\"", 400},
		{"wayyyyyyyyyyyyytooooloooonguuuserrnameeeeeeeeeeeeeeee", 400},
		{"     tuommii   ", 400},
		{"                  miikka                                       ", 200},
		{";miikka;", 400},
		{"jack_bauer", 200},
	}

	for _, item := range tt {
		rr, err := RequestCreateUser(UserInput{
			Username:  item.username,
			FirstName: "Example",
		}, rabbitCh)
		if err != nil {
			t.Error("Error while creating user", err)
		}
		if rr.Code != item.status {
			t.Error("Wrong status code", rr.Code, "Got that with", item.status, item.username)
		}
	}

	// Wait events to be generated
	time.Sleep(time.Second * 2)

	coll := store.GetClient().Database(consts.DatabaseName).Collection(consts.CollectionEvents)
	var eventsArr []events.Event
	ctx := context.Background()
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		t.Fatal(err.Error())
	}
	err = cursor.All(ctx, &eventsArr)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(eventsArr) != 3 {
		t.Fatal("events were not generated")
	}
}

func TestHandleGetUserByUsername(t *testing.T) {
	teardown := store.SetupTest(t)
	defer teardown()

	createExampleUsers(t)

	tt := []struct {
		routeVariable string
		status        int
	}{
		{"user1", 200},
		{"user2", 200},
		{"notexisting", 404},
	}

	for _, item := range tt {
		path := fmt.Sprintf("/users/%s", item.routeVariable)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatal(err)
		}
		router := mux.NewRouter()
		router.HandleFunc("/users/{username}", HandleGetUserByUsername)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		if status := recorder.Code; status != item.status {
			t.Error("expected", item.status, "got", status, "with", item.routeVariable, string(recorder.Body.String()))
		}
	}
}

func TestStoreGetMultipleUsers(t *testing.T) {
	teardown := store.SetupTest(t)
	defer teardown()

	createExampleUsers(t)

	exampleUsers := []string{"user1", "user2", "user3"}
	users, err := StoreGetUsersByUsername(exampleUsers)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != len(exampleUsers) {
		t.Error("Wrong amount of users")
	}
}

// user1 ... user4
func createExampleUsers(t *testing.T) {
	tt := []struct {
		username string
		status   int
	}{
		{"user1", 200},
		{"user2", 200},
		{"user3", 200},
		{"user4", 200},
	}
	for _, item := range tt {
		rr, err := RequestCreateUser(UserInput{
			Username:  item.username,
			FirstName: "Example",
		}, nil)
		if err != nil {
			t.Error("Error while creating user", err)
		}
		if rr.Code != item.status {
			t.Error("Wrong status code", rr.Code)
		}
	}
}
