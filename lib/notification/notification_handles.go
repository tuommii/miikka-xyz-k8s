package notification

import (
	"net/http"

	"github.com/streadway/amqp"
)

func HandleGetNotifications(ch *amqp.Channel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "notifications not yet implemented\n", 500)
	}
}
