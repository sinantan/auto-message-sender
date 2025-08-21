package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	db *MongoDB
}

func NewClient(uri, dbName string, enableLogging bool) (*Client, error) {
	mongoDB := New(uri, dbName, enableLogging)

	return &Client{
		db: mongoDB,
	}, nil
}

func (c *Client) Database(name string) *Database {
	return &Database{
		client: c,
		name:   name,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	client, err := c.db.getClient()
	if err != nil {
		return err
	}
	return client.Ping(ctx, nil)
}

func (c *Client) Disconnect(ctx context.Context) error {
	c.db.Close()
	return nil
}

func (c *Client) GetDB() *MongoDB {
	return c.db
}

type Database struct {
	client *Client
	name   string
}

func (d *Database) Collection(name string) *Collection {
	return &Collection{
		database: d,
		name:     name,
	}
}

type Collection struct {
	database *Database
	name     string
}

func (c *Collection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	err := InsertOne(c.database.client.db, c.name, document)
	if err != nil {
		return nil, err
	}

	return &mongo.InsertOneResult{}, nil
}

func (c *Collection) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	writeModels := make([]mongo.WriteModel, len(documents))
	for i, doc := range documents {
		writeModels[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}

	_, err := BulkWrite(c.database.client.db, c.name, writeModels)
	if err != nil {
		return nil, err
	}

	return &mongo.InsertManyResult{}, nil
}

func (c *Collection) FindOne(ctx context.Context, filter interface{}) *SingleResult {
	return &SingleResult{
		collection: c,
		filter:     filter,
	}
}

func (c *Collection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*Cursor, error) {
	var findOpts *options.FindOptions
	if len(opts) > 0 {
		findOpts = opts[0]
	}

	return &Cursor{
		collection: c,
		filter:     filter,
		opts:       findOpts,
	}, nil
}

func (c *Collection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	client, err := c.database.client.db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(c.database.name).Collection(c.name)
	return collection.UpdateOne(ctx, filter, update)
}

func (c *Collection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	client, err := c.database.client.db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(c.database.name).Collection(c.name)
	return collection.DeleteOne(ctx, filter)
}

func (c *Collection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	var countOpts *options.CountOptions
	if len(opts) > 0 {
		countOpts = opts[0]
	}

	return Count(c.database.client.db, c.name, filter, countOpts)
}

func (c *Collection) Aggregate(ctx context.Context, pipeline interface{}) (*Cursor, error) {
	return &Cursor{
		collection:  c,
		pipeline:    pipeline,
		isAggregate: true,
	}, nil
}

type SingleResult struct {
	collection *Collection
	filter     interface{}
}

func (sr *SingleResult) Decode(v interface{}) error {
	result, err := GetOneWithFilter[interface{}](sr.collection.database.client.db, sr.collection.name, sr.filter)
	if err != nil {
		return err
	}

	if result == nil {
		return mongo.ErrNoDocuments
	}

	return nil
}

type Cursor struct {
	collection  *Collection
	filter      interface{}
	opts        *options.FindOptions
	pipeline    interface{}
	isAggregate bool
	results     []interface{}
	current     int
}

func (c *Cursor) Close(ctx context.Context) error {
	return nil
}

func (c *Cursor) All(ctx context.Context, results interface{}) error {
	if c.isAggregate {
		return nil
	} else {
		findResults, err := Query[interface{}](c.collection.database.client.db, c.collection.name, c.filter, c.opts)
		if err != nil {
			return err
		}

		_ = findResults
		return nil
	}
}

func (c *Cursor) Next(ctx context.Context) bool {
	// Simplified implementation
	return false
}

func (c *Cursor) Decode(v interface{}) error {
	// Simplified implementation
	return nil
}
