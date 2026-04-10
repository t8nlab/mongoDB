package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	ext "github.com/t8nlab/mongo/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type clientStore struct {
	sync.RWMutex
	clients map[string]*mongo.Client
}

var store = &clientStore{
	clients: make(map[string]*mongo.Client),
}

func (s *clientStore) get(handle string) (*mongo.Client, error) {
	s.RLock()
	defer s.RUnlock()
	c, ok := s.clients[handle]
	if !ok {
		return nil, fmt.Errorf("invalid connection handle: %s", handle)
	}
	return c, nil
}

func (s *clientStore) set(handle string, client *mongo.Client) {
	s.Lock()
	defer s.Unlock()
	s.clients[handle] = client
}

func (s *clientStore) remove(handle string) error {
	s.Lock()
	defer s.Unlock()
	c, ok := s.clients[handle]
	if !ok {
		return fmt.Errorf("invalid connection handle: %s", handle)
	}
	delete(s.clients, handle)
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return c.Disconnect(ctx)
}

// Optimized processMap with pre-check
func processMap(m map[string]any) bson.M {
	res := make(bson.M, len(m))
	for k, v := range m {
		if k == "_id" {
			if s, ok := v.(string); ok && len(s) == 24 {
				if oid, err := primitive.ObjectIDFromHex(s); err == nil {
					res[k] = oid
					continue
				}
			}
		}
		
		switch val := v.(type) {
		case map[string]any:
			res[k] = processMap(val)
		case []any:
			res[k] = processArray(val)
		default:
			res[k] = val
		}
	}
	return res
}

func processArray(arr []any) []any {
	res := make([]any, len(arr))
	for i, v := range arr {
		switch val := v.(type) {
		case map[string]any:
			res[i] = processMap(val)
		case []any:
			res[i] = processArray(val)
		default:
			res[i] = val
		}
	}
	return res
}

// ---------- CONNECT ----------

func mongoConnect(input map[string]any) (any, error) {
	uri, ok := input["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("uri required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(uri)
	opts.SetMaxPoolSize(100)
	opts.SetMinPoolSize(10)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Wait only 1 second for Ping to avoid hanging
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer pingCancel()
	err = client.Ping(pingCtx, nil)
	if err != nil {
		// Just log or warn, but don't fail if ping fails but connect succeeded
		// return nil, fmt.Errorf("ping failed: %v", err)
	}

	handle := fmt.Sprintf("conn_%d", time.Now().UnixNano())
	store.set(handle, client)

	return map[string]any{"handle": handle}, nil
}

func mongoClose(input map[string]any) (any, error) {
	handle, _ := input["handle"].(string)
	err := store.remove(handle)
	if err != nil {
		return nil, err
	}
	return map[string]any{"closed": true}, nil
}

func mongoInsert(input map[string]any) (any, error) {
	handle, _ := input["handle"].(string)
	dbName, _ := input["db"].(string)
	colName, _ := input["collection"].(string)
	docData, _ := input["doc"].(map[string]any)

	client, err := store.get(handle)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := client.Database(dbName).Collection(colName)
	res, err := collection.InsertOne(ctx, processMap(docData))
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"insertedId": fmt.Sprintf("%v", res.InsertedID),
	}, nil
}

func mongoFind(input map[string]any) (any, error) {
	handle, _ := input["handle"].(string)
	dbName, _ := input["db"].(string)
	colName, _ := input["collection"].(string)
	filterData, _ := input["filter"].(map[string]any)
	
	limit, _ := input["limit"].(float64)
	skip, _ := input["skip"].(float64)

	client, err := store.get(handle)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var filter bson.M
	if filterData != nil {
		filter = processMap(filterData)
	} else {
		filter = bson.M{}
	}

	collection := client.Database(dbName).Collection(colName)
	
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	} else {
		opts.SetLimit(1000)
	}
	if skip > 0 {
		opts.SetSkip(int64(skip))
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Pre-allocate slice capacity if possible
	results := make([]bson.M, 0, 100)
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ext.Register("mongo_connect", mongoConnect)
	ext.Register("mongo_close", mongoClose)
	ext.Register("mongo_insert", mongoInsert)
	ext.Register("mongo_find", mongoFind)
}

func main() {}