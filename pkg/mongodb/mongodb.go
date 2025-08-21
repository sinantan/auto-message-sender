package mongodb

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"runtime"
	"time"
)

//var db *MongoDB

type MongoDB struct {
	ConnectionString string
	DBName           string
	Client           *mongo.Client
	isLoggingEnabled bool
}

func New(connectionString string, dbName string, loggingEnabled bool) *MongoDB {
	return &MongoDB{
		ConnectionString: connectionString,
		DBName:           dbName,
		Client:           nil,
		isLoggingEnabled: loggingEnabled,
	}
}

func (m *MongoDB) logMongo() {
	if !m.isLoggingEnabled {
		return
	}

	callers := make([]string, 0)

	for i := 3; i >= 1; i-- {
		if pc, _, _, ok := runtime.Caller(i); !ok {
			fmt.Println("Unable to retrieve caller information")
			return
		} else {
			cf := runtime.FuncForPC(pc).Name()
			cf = runtime.FuncForPC(pc).Name()
			callers = append(callers, cf)
		}
	}

	finalLogMessage := ""
	for index, caller := range callers {
		finalLogMessage += caller
		if index != len(callers)-1 {
			finalLogMessage += " -- >"
		}
	}

	log.Println(fmt.Sprintf("MONGODB REQUEST: %s", finalLogMessage))
}

func (m *MongoDB) GetClient() (*mongo.Client, error) {
	return m.getClient()
}

func (m *MongoDB) getClient() (*mongo.Client, error) {
	if m.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := m.Client.Ping(ctx, nil)
		if err == nil {
			return m.Client, nil
		}
		m.Client.Disconnect(context.Background())
		m.Client = nil
	}

	m.logMongo()

	var err error
	clientOptions := options.Client().ApplyURI(m.ConnectionString).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(100).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		m.Client, err = mongo.Connect(ctx, clientOptions)
		cancel()

		if err == nil {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			err = m.Client.Ping(ctx, nil)
			cancel()

			if err == nil {
				return m.Client, nil
			}
		}

		if m.Client != nil {
			m.Client.Disconnect(context.Background())
			m.Client = nil
		}

		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to connect to MongoDB after %d attempts: %v", maxRetries, err)
}

func (m *MongoDB) Close() {
	if m.Client == nil {
		return
	}

	err := m.Client.Disconnect(context.TODO())
	if err != nil {
		log.Println(err)
	}

	m.Client = nil
}

func GetOneById[T any](db *MongoDB, collectionName string, id string) (*T, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	filter := bson.D{{"_id", id}}

	var result *T
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return result, err
}

func GetOneWithFilter[T any](db *MongoDB, collectionName string, filter interface{}) (*T, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var result *T
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return result, err
}

func InsertOne[T any](db *MongoDB, collectionName string, record T) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	_, err = collection.InsertOne(ctx, record)
	return err
}

func UpdateOne[T any](db *MongoDB, collectionName string, id string, record *T) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	collection := client.Database(db.DBName).Collection(collectionName)
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", record}}
	opts := options.Update() //.SetUpsert(true)

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	_, err = collection.UpdateOne(ctx, filter, update, opts) //result
	if err != nil {
		return err
	}

	return nil
}

func IncrementValue(db *MongoDB, collectionName string, id string, fieldName string, value int) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	update := bson.M{"$inc": bson.M{fieldName: value}}

	collection := client.Database(db.DBName).Collection(collectionName)

	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
	if err != nil {
		return err
	}

	return err
}

func SetValue(db *MongoDB, collectionName string, id string, fieldName string, value any) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{fieldName: value}}

	collection := client.Database(db.DBName).Collection(collectionName)

	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
	if err != nil {
		return err
	}

	return err
}

func UpsertOne[T any](db *MongoDB, collectionName string, id string, record *T) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	collection := client.Database(db.DBName).Collection(collectionName)
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", record}}
	opts := options.Update().SetUpsert(true)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	//defer client.Disconnect(ctx)

	_, err = collection.UpdateOne(ctx, filter, update, opts) //result
	if err != nil {
		return err
	}

	return nil
}

func BulkWrite(db *MongoDB, collectionName string, writeModels []mongo.WriteModel) (int, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return 0, err
	}

	bulkWriteOptions := options.BulkWrite().SetOrdered(false)

	collection := client.Database(db.DBName).Collection(collectionName)

	numberOfDuplicates := 0

	_, err = collection.BulkWrite(context.TODO(), writeModels, bulkWriteOptions)
	if err != nil {
		if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
			for _, writeError := range bulkWriteException.WriteErrors {
				if writeError.Code == 11000 {
					numberOfDuplicates++
				} else {
					return numberOfDuplicates, err
				}
			}
		} else {
			return numberOfDuplicates, err
		}
	}

	return numberOfDuplicates, nil
}

func BulkUpdate(db *MongoDB, collectionName string, updateModels []mongo.WriteModel) (int, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return 0, err
	}

	bulkWriteOptions := options.BulkWrite().SetOrdered(false)

	collection := client.Database(db.DBName).Collection(collectionName)

	numberOfFailures := 0

	_, err = collection.BulkWrite(context.TODO(), updateModels, bulkWriteOptions)
	if err != nil {
		if bulkWriteException, ok := err.(mongo.BulkWriteException); ok {
			for _, _ = range bulkWriteException.WriteErrors {
				numberOfFailures++
			}
		} else {
			return numberOfFailures, err
		}
	}

	numberOfUpdates := len(updateModels) - numberOfFailures

	return numberOfUpdates, nil
}

func Query[T any](db *MongoDB, collectionName string, filter interface{}, opts *options.FindOptions) ([]T, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	//defer cancel()
	//defer client.Disconnect(ctx)

	//filter := bson.D{{"syncstatus.lastreleasecheckdate", nil}}
	//opts := options.Find().SetSort(bson.D{{"credate", -1}})

	if cursor, err := collection.Find(ctx, filter, opts); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	} else {
		var result []T
		if err = cursor.All(ctx, &result); err != nil {
			return nil, err
		}

		if result == nil {
			return []T{}, nil
		}

		return result, nil
	}
}

func GetAll[T any](db *MongoDB, collectionName string) ([]T, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(db.DBName).Collection(collectionName)
	filter := bson.D{}

	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []T
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteOne(db *MongoDB, collectionName string, id string) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	filter := bson.D{{"_id", id}}

	_, err = collection.DeleteOne(ctx, filter)
	return err
}

func DeleteAll(db *MongoDB, collectionName string, filter interface{}) error {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	_, err = collection.DeleteMany(ctx, filter)
	return err
}

func Count(db *MongoDB, collectionName string, filter interface{}, opts *options.CountOptions) (int64, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return -1, err
	}

	collection := client.Database(db.DBName).Collection(collectionName)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	if count, err := collection.CountDocuments(ctx, filter, opts); err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}

		return -1, err
	} else {
		return count, nil
	}
}

func Aggregate[T any](db *MongoDB, collectionName string, pipeline interface{}) ([]T, error) {
	db.logMongo()
	client, err := db.getClient()
	if err != nil {
		return nil, err
	}

	collection := client.Database(db.DBName).Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 300*time.Second) // Increased to 5 minutes

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	var results []T
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		return []T{}, nil
	}

	return results, nil
}
