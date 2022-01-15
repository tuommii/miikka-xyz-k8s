# Spaces are not allowed!
VERSION=$(shell cat VERSION)
DATE=$(shell date +%d.%m.%Y)
COMMIT=$(shell git rev-parse --short=8 HEAD)

whole:
	docker-compose --file docker-compose-all.yml up

test:
	go test -failfast -v -count=1 -p 1 ./...

swagger:
	swag init -g server/server.go

# Build binaries for events, api and traffic job
# TODO: Not working in docker file yet!
build:
	go build -o bin/devops-events -trimpath -ldflags \
	'-X miikka.xyz/devops-app/consts.Build=$(DATE) -X miikka.xyz/devops-app/consts.Version=$(VERSION) -X miikka.xyz/devops-app/consts.Commit=$(COMMIT)'\
	 cmd/events/*.go

	go build -o bin/devops-api -trimpath -ldflags \
	'-X miikka.xyz/devops-app/consts.Build=$(DATE) -X miikka.xyz/devops-app/consts.Version=$(VERSION) -X miikka.xyz/devops-app/consts.Commit=$(COMMIT)'\
	 cmd/api/*.go

	go build -o bin/traffic-job -trimpath -ldflags \
	'-X miikka.xyz/devops-app/consts.Build=$(DATE) -X miikka.xyz/devops-app/consts.Version=$(VERSION) -X miikka.xyz/devops-app/consts.Commit=$(COMMIT)'\
	 cmd/traffic_job/*.go

clean:
	rm -rf bin/
