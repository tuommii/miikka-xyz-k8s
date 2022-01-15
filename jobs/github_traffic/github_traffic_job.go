package github_traffic

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/go-github/v41/github"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/store"
)

// client holds global Github API client. Using global so every job doesn't need to create a client
// and this job might be started with Cron and UI
var client *github.Client

// WorkerResult represents data that each worker sends to the results channel
type WorkerResult struct {
	RepoName     string
	TrafficViews *github.TrafficViews
}

// init Github API client in package's init function
func init() {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_TOKEN")},
	)
	tokenClient := oauth2.NewClient(context.Background(), tokenSource)
	client = github.NewClient(tokenClient)
	client.Client().Timeout = time.Second * 45
}

// DoGithubTrafficStats will get user's traffic data (visitor counts) from GitHub and saves those
// to database
func DoGithubTrafficStats() error {
	// How many repositories will be retrieved at once.
	// Those will be then splitted for each worker.
	const pageSize = 100
	// This will get increased atomically when error happens
	var errorHasOccured int64 = 0

	// Each worker sends result to this channel
	resultsCh := make(chan *WorkerResult)
	doneCh := make(chan bool)

	// All workers will be canceled when first error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go listenResultsChannel(ctx, cancel, resultsCh, doneCh, &errorHasOccured)

	// Only repositories where user is a owner, max pageSize at time
	options := &github.RepositoryListOptions{Affiliation: "owner", ListOptions: github.ListOptions{PerPage: pageSize}}

	// Until all pages has been fetched
	i := 1
	for {
		// Don't fetch more repositories if worker has failed
		if errorHasOccured > 0 {
			log.Println("cancel job, error has occured")
			close(resultsCh)
			break
		}

		repos, resp, err := client.Repositories.List(ctx, "", options)
		if err != nil {
			return err
		}
		log.Println("page", i, "has", len(repos), "repositories")

		// Save repository data like repo URL to different collection
		err = saveRepositoryData(ctx, repos, &errorHasOccured)
		if err != nil {
			log.Println(err)
			atomic.AddInt64(&errorHasOccured, 1)
			cancel()
			return err
		}

		var workerWG sync.WaitGroup
		workerWG.Add(1)
		launchWorkers(ctx, cancel, resultsCh, &workerWG, repos, &errorHasOccured)
		// Don't start new workers before previously started are done
		workerWG.Wait()

		// All pages fetched
		if resp.NextPage == 0 {
			log.Println("this was last page, close results channel")
			close(resultsCh)
			break
		}

		// Otherwise fetch next page
		options.ListOptions.Page = resp.NextPage
		i++
	}

	<-doneCh

	if errorHasOccured > 0 {
		return errors.New("job failed")
	}

	return nil
}

// saveResultParams holds parameters for saveResults-function
type saveResultParams struct {
	view            *github.TrafficData
	coll            *mongo.Collection
	operations      *[]mongo.WriteModel
	i               *int
	repoName        string
	saveAtOnceCount int
	errorHasOccured *int64
}

// listenResultsChannel collects results from workers
func listenResultsChannel(ctx context.Context, cancel func(), resultsCh chan *WorkerResult, doneCh chan bool, errorHasOccured *int64) {
	defer func() {
		doneCh <- true
		close(doneCh)
	}()

	var i int
	params := saveResultParams{
		coll:            store.GetClient().Database(consts.DatabaseName).Collection(consts.CollectionRepoTraffic),
		operations:      &[]mongo.WriteModel{},
		saveAtOnceCount: 100,
		i:               &i,
	}

	// Loop results
	for workerResult := range resultsCh {
		if workerResult == nil {
			continue
		}

		// Each result has array of views
		for _, view := range workerResult.TrafficViews.Views {
			params.view = view
			params.repoName = workerResult.RepoName
			err := saveResults(ctx, cancel, params)
			if err != nil {
				log.Println(err)
				atomic.AddInt64(params.errorHasOccured, 1)
				cancel()
				return
			}
		}
	}

	// All results has been saved
	if *params.i == 0 {
		return
	}

	// Save rest of the results
	_, err := params.coll.BulkWrite(ctx, *params.operations, &options.BulkWriteOptions{})
	if err != nil {
		log.Println(err)
		atomic.AddInt64(errorHasOccured, 1)
		cancel()
		return
	}
}

// saveResults collects results and saves those to database when amount of saveAtOnce is exceeded
// TODO: Transactions, replica mode in docker?
func saveResults(ctx context.Context, cancel func(), params saveResultParams) error {
	log.Println(params.repoName, params.view.Timestamp.String()[:10], *params.view.Count, *params.view.Uniques)

	nameAndTimestampFilter := []bson.M{
		{"name": params.repoName},
		{"timestamp": params.view.Timestamp.Time},
	}
	filter := bson.M{"$and": nameAndTimestampFilter}

	update := bson.M{
		"$set": bson.M{
			"name":         params.repoName,
			"views":        *params.view.Count,
			"unique_views": *params.view.Uniques,
			"timestamp":    params.view.Timestamp.Time,
		},
	}

	updateModel := mongo.NewUpdateOneModel()
	updateModel.SetFilter(filter)
	updateModel.SetUpdate(update)
	updateModel.SetUpsert(true)
	*params.operations = append(*params.operations, updateModel)

	*params.i++
	if *params.i >= params.saveAtOnceCount {
		_, err := params.coll.BulkWrite(ctx, *params.operations, &options.BulkWriteOptions{})
		if err != nil {
			return err
		}
		*params.i = 0
		*params.operations = []mongo.WriteModel{}
	}
	return nil
}

// launchWorkers start's goroutines for each worker
func launchWorkers(ctx context.Context, cancel func(), resultsCh chan *WorkerResult, wg *sync.WaitGroup, chunk []*github.Repository, errorHasOccured *int64) {
	defer wg.Done()
	const workersCount = 2
	log.Println("workers count", workersCount)

	// Split array equally to each worker, last one takes extra ones
	chunkLen := len(chunk)
	amountPerWorker := chunkLen / workersCount

	for i := 0; i < workersCount; i++ {
		from := i * amountPerWorker
		to := (i + 1) * amountPerWorker
		wg.Add(1)

		// Last worker takes care of extra ones
		isLastWorker := (i+1 >= workersCount)
		if isLastWorker {
			go runWorker(ctx, cancel, resultsCh, wg, errorHasOccured, chunk[from:], i+1)
		} else {
			go runWorker(ctx, cancel, resultsCh, wg, errorHasOccured, chunk[from:to], i+1)
		}
	}
}

// runWorker runs task for each repository in chunk
func runWorker(ctx context.Context, cancel func(), resultsCh chan *WorkerResult, wg *sync.WaitGroup, errorHasOccured *int64, chunk []*github.Repository, index int) {
	defer func() {
		log.Println("worker", index, "- done!")
		wg.Done()
	}()
	log.Println("started worker", index, "with", len(chunk), "elements")

	for _, elem := range chunk {
		// Before doing next task, check if context is already canceled / other worker has failed.
		select {
		case <-ctx.Done():
			log.Println("some worker has failed, shutting down worker", index)
			atomic.AddInt64(errorHasOccured, 1)
			return
		default:
			// Actual task
			runWorkerTask(ctx, cancel, resultsCh, elem, index, errorHasOccured)
		}
	}

}

// runWorkerTask does the actual task
func runWorkerTask(ctx context.Context, cancel func(), resultsCh chan *WorkerResult, repo *github.Repository, index int, errorHasOccured *int64) {
	// Fetch views for repo
	views, _, err := client.Repositories.ListTrafficViews(ctx, "tuommii", *repo.Name, &github.TrafficBreakdownOptions{})
	if err != nil {
		log.Println(err)
		atomic.AddInt64(errorHasOccured, 1)
		cancel()
		return
	}

	// Send data to channel
	res := &WorkerResult{
		RepoName:     *repo.Name,
		TrafficViews: views,
	}
	resultsCh <- res
}

func saveRepositoryData(ctx context.Context, repos []*github.Repository, errorHasOccured *int64) error {
	operations := make([]mongo.WriteModel, 0)
	for _, r := range repos {
		filter := bson.M{"name": r.Name}
		update := bson.M{
			"$set": bson.M{
				"name": r.Name,
				"url":  r.GetHTMLURL(),
			}}
		updateModel := mongo.NewUpdateOneModel()
		updateModel.SetFilter(filter)
		updateModel.SetUpdate(update)
		updateModel.SetUpsert(true)
		operations = append(operations, updateModel)
	}

	coll := store.GetClient().Database(consts.DatabaseName).Collection(consts.CollectionRepos)
	_, err := coll.BulkWrite(ctx, operations, &options.BulkWriteOptions{})
	if err != nil {
		return err
	}
	log.Println("saving repo data succeed")
	return nil
}
