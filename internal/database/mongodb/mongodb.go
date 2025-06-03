package mongodb

import (
	"context"
	"fmt"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents a MongoDB database connection
type MongoDB struct {
	config types.Config
	client *mongo.Client
	db     *mongo.Database
}

// New creates a new MongoDB database connection
func NewMongoDBDatabase(config types.Config) (*MongoDB, error) {
	return &MongoDB{
		config: config,
	}, nil
}

// Connect establishes a connection to the MongoDB database
func (m *MongoDB) Connect(ctx context.Context) error {
	// Set client options
	clientOptions := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%d",
			m.config.Username,
			m.config.Password,
			m.config.Host,
			m.config.Port,
		)).
		SetMaxPoolSize(uint64(m.config.MaxOpenConns)).
		SetMinPoolSize(uint64(m.config.MaxIdleConns)).
		SetMaxConnIdleTime(m.config.ConnMaxIdleTime)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.db = client.Database(m.config.Database)
	return nil
}

// Close closes the MongoDB database connection
func (m *MongoDB) Close() error {
	if m.client != nil {
		return m.client.Disconnect(context.Background())
	}
	return nil
}

// CreateTable creates a collection in MongoDB
func (m *MongoDB) CreateTable(ctx context.Context, tableName string, columns []types.Column) error {
	// In MongoDB, we don't need to explicitly create collections
	// They are created automatically when we insert data
	return nil
}

// InsertData inserts data into a MongoDB collection
func (m *MongoDB) InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Convert data to BSON documents
	documents := make([]interface{}, len(data))
	for i, row := range data {
		documents[i] = row
	}

	// Insert documents
	_, err := m.db.Collection(tableName).InsertMany(ctx, documents)
	return err
}

// BeginTransaction starts a new MongoDB transaction
func (m *MongoDB) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	session, err := m.client.StartSession()
	if err != nil {
		return nil, err
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return nil, err
	}

	return &MongoDBTransaction{
		session: session,
		ctx:     ctx,
	}, nil
}

// GetDriver returns the MongoDB driver name
func (m *MongoDB) GetDriver() string {
	return "mongodb"
}

// CreateIndex creates an index on a MongoDB collection
func (m *MongoDB) CreateIndex(ctx context.Context, tableName string, index types.Index) error {
	if len(index.Columns) == 0 {
		return fmt.Errorf("no columns specified for index")
	}

	// Create index model
	keys := make(bson.D, len(index.Columns))
	for i, col := range index.Columns {
		keys[i] = bson.E{Key: col, Value: 1}
	}

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(index.Name),
	}

	// Set index options
	if index.IsUnique {
		indexModel.Options.SetUnique(true)
	}

	// Set additional properties
	if len(index.Properties) > 0 {
		for key, value := range index.Properties {
			switch key {
			case "sparse":
				if sparse, ok := value.(bool); ok {
					indexModel.Options.SetSparse(sparse)
				}
			case "expireAfterSeconds":
				if ttl, ok := value.(int32); ok {
					indexModel.Options.SetExpireAfterSeconds(ttl)
				}
			case "background":
				if background, ok := value.(bool); ok {
					indexModel.Options.SetBackground(background)
				}
			}
		}
	}

	// Create the index
	_, err := m.db.Collection(tableName).Indexes().CreateOne(ctx, indexModel)
	return err
}

// MongoDBTransaction represents a MongoDB transaction
type MongoDBTransaction struct {
	session mongo.Session
	ctx     context.Context
}

// Commit commits the MongoDB transaction
func (t *MongoDBTransaction) Commit() error {
	return t.session.CommitTransaction(t.ctx)
}

// Rollback rolls back the MongoDB transaction
func (t *MongoDBTransaction) Rollback() error {
	return t.session.AbortTransaction(t.ctx)
}
