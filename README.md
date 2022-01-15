## Introduction

My over-engineered homepage project to get an idea of the Kubernetes. It shows traffic data from all my GitHub repositories. I added RabbitMQ, Redis and MongoDB so I got to play around with multiple k8s resources. Traffic data is fetched concurrently and
Kubernetes manages that job as a cron job. Also now my old hobby projects are under Kubernetes.

## See it in action
Go to https://miikka.xyz or run the project locally with your own `GITHUB_API_TOKEN=token_here`. Fetching traffic data automatically on start up, requires `RUN_JOBS_ON_STARTUP=true` to be set.

**Docker reads environment variables from `.env` file, but code doesn't.**

Run infrastructure and services with:
```
docker-compose --file docker-compose-all.yml up --build
```

Go to http://localhost


## Development

### Project structure

| Package  | Description |
| ------------- | ------------- |
| `assets/k8s` | Kubernetes manifests. Currently not in GitHub repository |
| `cache` | Redis related code |
| `cmd` | Contains main.go files for binaries |
| `consts` | Constants that are being used in multiple places |
| `docs` | Generated Swagger files, just as an example |
| `events` | RabbitMQ related code |
| `jobs` | Job for fetching traffic data from GitHub API concurrently. Kubernetes runs this as a cron job |
| `lib` | Contains models, routes, store functions and tests for each entity |
| `lib/user` | Just an example, no use for this project |
| `server` | Server setup |
| `server/tmpls` | Contains HTML-template which will be injected to binary |
| `store` | Database connection |

### Local development
Run infrastructure with:
```
docker-compose up -d
```

Build binaries with `make build` and run those or use VS Code's `.vscode/launch.json`

### Test

WIP.

If all tests are to be run at the same time, parallel count must be set to 1, because
every test cleans database. Cached test will be disabled with that `-count=1`
```
go test -v -count=1 -p 1 ./...
```

Run one test function
```
go test -v -timeout 180s -count=1 -run ^TestJob$ miikka.xyz/devops-app/lib/job
```

## Deploy application
Build Docker images and push to Docker registry

Eventlistener
```
docker build . -t tuommii/miikka-xyz-events -f Dockerfile-eventlistener
docker push tuommii/miikka-xyz-events
```

Traffic job
```
docker build . -t tuommii/miikka-xyz-traffic-job -f Dockerfile-traffic-job
docker push tuommii/miikka-xyz-traffic-job
```

API
```
docker build . -t tuommii/miikka-xyz-api -f Dockerfile-api
docker push tuommii/miikka-xyz-api
```

## Update k8s resources
In case updating Kubernetes resources is needed, run following commands
```
kubectl apply -f assets/k8s/events.yml
kubectl apply -f assets/k8s/traffic-job.yml
kubectl apply -f assets/k8s/api.yml
```


## Notes

### Swagger
Swagger URL is http://localhost:8080/swagger/ if enabled

Update docs with `make swagger`

### k8s
Get pods
```
kubectl get pod
```

Verify service
```
kubectl get svc miikka-xyz-api
```

Get a shell to the running container
```
kubectl exec --stdin --tty mongo-container-name -- /bin/bash
```

Get a YAML file
```
kubectl get deploy rabbitmq -o yaml > rabbit.yaml
```

### Docker
Delete all volumes
```
docker volume rm $(docker volume ls -q)
```

## Links
[How To Set Up an Nginx Ingress on DigitalOcean Kubernetes Using Helm](https://www.digitalocean.com/community/tutorials/how-to-set-up-an-nginx-ingress-on-digitalocean-kubernetes-using-helm)

[How to Set Up an Nginx Ingress with Cert-Manager on DigitalOcean Kubernetes](https://www.digitalocean.com/community/tutorials/how-to-set-up-an-nginx-ingress-with-cert-manager-on-digitalocean-kubernetes)
