package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func RequestCreateUser(user UserInput, ch *amqp.Channel) (*httptest.ResponseRecorder, error) {
	bytesJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(bytesJson))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	router.HandleFunc("/user", HandleCreateUser(ch))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder, nil
}
