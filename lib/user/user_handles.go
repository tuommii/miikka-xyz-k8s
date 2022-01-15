package user

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/events"
	"miikka.xyz/devops-app/utils"
)

func HandleCreateUser(ch *amqp.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read request body
		bytes, err := ioutil.ReadAll(io.LimitReader(r.Body, consts.MaxBodySizeBytes))
		if err != nil {
			log.Println(err)
			http.Error(w, consts.ErrBody, http.StatusInternalServerError)
			return
		}

		// JSON to UserInput
		userInput := UserInput{}
		err = json.Unmarshal(bytes, &userInput)
		if err != nil {
			log.Println(err)
			http.Error(w, consts.ErrJSON, http.StatusBadRequest)
			return
		}

		// Trim username
		username := strings.TrimSpace(userInput.Username)
		if username == "" || len(username) > 32 || !utils.OnlyAlphaNumberOrUnderscore(username) {
			log.Println("invalid username", username)
			http.Error(w, "invalid username", http.StatusBadRequest)
			return
		}
		log.Printf("username after trimming: [%s]\n", username)

		// Check is username already taken
		userFound, err := StoreGetUserByUsername(username)
		if err != nil {
			log.Println(err)
			http.Error(w, consts.ErrDatabase, http.StatusInternalServerError)
			return
		}
		if userFound != nil {
			msg := fmt.Sprintf(consts.ErrAlreadyExists, username)
			log.Println(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		// Create new user
		id, err := StoreCreateUser(&userInput)
		if err != nil {
			log.Println(err)
			http.Error(w, consts.ErrDatabase, http.StatusInternalServerError)
			return
		}

		// Publish user created event in background
		// Some tests might use nil value so check it
		if ch != nil {
			go func() {
				event := events.Event{
					CreatedAt: time.Now(),
					ObjectID:  id,
					Type:      consts.EventUserCreated,
				}
				err = events.Publish(ch, &event)
				if err != nil {
					log.Println(err)
				}
			}()
		}

		fmt.Fprintln(w, id.Hex())
	}
}

// HandleGetUserByUsername
func HandleGetUserByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		http.Error(w, consts.ErrNotFound, http.StatusNotFound)
		return
	}

	user, err := StoreGetUserByUsername(username)
	if err != nil {
		log.Println(err)
		http.Error(w, consts.ErrDatabase, http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Println(err)
		http.Error(w, consts.ErrNotFound, http.StatusNotFound)
		return
	}
	fmt.Fprintln(w, user.Username)
}
