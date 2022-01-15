package repo

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/store"
)

type TrafficData struct {
	RepositoryName string    `bson:"name" json:"name"`
	Views          int       `bson:"views" json:"views"`
	UniqueViews    int       `bson:"unique_views" json:"unique_views"`
	Timestamp      time.Time `bson:"timestamp" json:"timestamp"`
	// $lookup
	RepositoryData RepositoryData `bson:"_meta" json:"_meta"`
}

type RepositoryData struct {
	URL string `bson:"url" json:"url"`
}

// ReposByNameMap type will be saved to Redis
type ReposByNameMap map[string][]TrafficData

// MarshalBinary implements Marshaler interface so this type can be saved to Redis
func (r ReposByNameMap) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(r)
	return data, err
}

func StoreGetRepositoryTraffic(ctx context.Context, since time.Time) ([]TrafficData, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionRepoTraffic)

	cursor, err := coll.Find(ctx, bson.M{"timestamp": bson.M{"$gte": since}})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var repos []TrafficData
	err = cursor.All(ctx, &repos)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return repos, nil
}

func StoreGetRepositoryTrafficWithMeta(ctx context.Context, since time.Time) ([]TrafficData, error) {
	client := store.GetClient()
	coll := client.Database(consts.DatabaseName).Collection(consts.CollectionRepoTraffic)

	pipe := make([]bson.M, 0)
	match := bson.M{"$match": bson.M{"timestamp": bson.M{"$gte": since}}}
	lookup := bson.M{"$lookup": bson.M{
		"from":         consts.CollectionRepos,
		"localField":   "name",
		"foreignField": "name",
		"as":           "_meta",
	}}
	// Transform repository data to single object, not as array
	unwind := bson.M{"$unwind": "$_meta"}

	pipe = append(pipe, match)
	pipe = append(pipe, lookup)
	pipe = append(pipe, unwind)
	pipe = append(pipe, bson.M{"$sort": bson.M{"timestamp": -1, "name": 1}})

	cursor, err := coll.Aggregate(ctx, pipe, nil)
	if err != nil {
		return nil, err
	}

	var repos []TrafficData
	err = cursor.All(ctx, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func FormatRepositorysToMap(repos []TrafficData) ReposByNameMap {
	reposByName := make(ReposByNameMap)
	for _, r := range repos {
		list, listFound := reposByName[r.RepositoryName]
		if !listFound {
			list = make([]TrafficData, 0)
		}
		list = append(list, r)
		reposByName[r.RepositoryName] = list
	}

	/*for _, traffic := range reposByName {
		sort.Slice(traffic, func(i, j int) bool {
			return traffic[i].Timestamp.Before(traffic[i].Timestamp)
		})
	}*/

	return reposByName
}

func TemplateGetLink(repo TrafficData) string {
	return repo.RepositoryData.URL
}
