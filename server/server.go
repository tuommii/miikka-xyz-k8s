package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"miikka.xyz/devops-app/cache"
	"miikka.xyz/devops-app/consts"
	_ "miikka.xyz/devops-app/docs"
	"miikka.xyz/devops-app/lib/notification"
	"miikka.xyz/devops-app/lib/repo"
	"miikka.xyz/devops-app/utils"
)

//go:embed tmpls/index.go.html
var embedFS embed.FS

type Server struct {
	HTTP         *http.Server
	EventChannel *amqp.Channel
	Cache        *cache.Cache
}

func New(port string, ch *amqp.Channel, cacheClient *cache.Cache) *Server {
	server := &Server{
		EventChannel: ch,
		Cache:        cacheClient,
		HTTP: &http.Server{
			Handler:           mux.NewRouter(),
			Addr:              "0.0.0.0:" + port,
			WriteTimeout:      30 * time.Second,
			ReadTimeout:       30 * time.Second,
			IdleTimeout:       30 * time.Second,
			ReadHeaderTimeout: 30 * time.Second,
		},
	}
	server.initRoutes()
	return server
}

// @title miikka.xyz API with Swagger
// @version 1.0
// @description Demo

// @contact.name Miikka Tuominen
// @contact.url https://miikka.xyz
// @contact.email tuommii@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:4242
// @BasePath /
// @query.collection.format multi

// @x-extension-openapi {"example": "value on a json format"}
func (s *Server) initRoutes() {
	router, ok := s.HTTP.Handler.(*mux.Router)
	if !ok {
		log.Fatal("Error with router")
	}

	// Uncomment to enable Swagger
	// router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/_health", healthCheck).Methods("GET")
	router.HandleFunc("/notification", notification.HandleGetNotifications(s.EventChannel)).Methods("GET")
	router.HandleFunc("/", s.home).Methods("GET")
}

// home renders template with traffic statistics
func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	templateFuncs := map[string]interface{}{
		"GetLink":        repo.TemplateGetLink,
		"DateToEuropean": utils.DateToEuropean,
	}
	tpl, err := template.New("home").Funcs(templateFuncs).ParseFS(embedFS, "tmpls/index.go.html")
	if err != nil {
		log.Println("error while creating template", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	templateData := map[string]interface{}{
		"version":   consts.Version,
		"buildTime": consts.Build,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	reposByName, err := s.Cache.GetTrafficData(ctx)

	if err != nil {
		log.Println("could not find data from cache")
		templateData["repos"] = make(repo.ReposByNameMap)
	} else {
		templateData["repos"] = reposByName
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tpl.ExecuteTemplate(w, "home", templateData); err != nil {
		log.Println(err)
		return
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	log.Println("health check")
	fmt.Fprintf(w, "OK")
}
