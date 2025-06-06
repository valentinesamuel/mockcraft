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

// CreateConstraint creates a constraint on a MongoDB collection
func (m *MongoDB) CreateConstraint(ctx context.Context, tableName string, constraint types.Constraint) error {
	if constraint.Type == "foreign_key" {
		// For MongoDB, we'll create an index on the foreign key fields
		// This helps with query performance and ensures uniqueness if needed
		keys := bson.D{}

		// Add each column to the index
		for _, col := range constraint.Columns {
			keys = append(keys, bson.E{Key: col, Value: 1})
		}

		indexModel := mongo.IndexModel{
			Keys: keys,
		}

		// Create the index
		_, err := m.db.Collection(tableName).Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			return fmt.Errorf("failed to create index for foreign key constraint: %w", err)
		}
	}
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
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	return nil
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

// DropTable drops a collection in MongoDB
func (m *MongoDB) DropTable(ctx context.Context, tableName string) error {
	return m.db.Collection(tableName).Drop(ctx)
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

// GetAllIDs retrieves all _id values from a collection
func (m *MongoDB) GetAllIDs(ctx context.Context, tableName string) ([]interface{}, error) {
	cursor, err := m.db.Collection(tableName).Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	ids := make([]interface{}, len(results))
	for i, doc := range results {
		ids[i] = doc["_id"]
	}
	return ids, nil
}

// GetAllForeignKeys retrieves all foreign key values from a collection
func (m *MongoDB) GetAllForeignKeys(ctx context.Context, tableName, columnName string) ([]interface{}, error) {
	collection := m.db.Collection(tableName)

	// Create a projection to only get the specified column
	opts := options.Find().SetProjection(bson.D{{Key: columnName, Value: 1}})

	// Find all documents
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents in %s: %w", tableName, err)
	}
	defer cursor.Close(ctx)

	// Decode the results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results from %s: %w", tableName, err)
	}

	// Extract the foreign key values
	values := make([]interface{}, 0, len(results))
	for _, doc := range results {
		if value, ok := doc[columnName]; ok {
			values = append(values, value)
		}
	}

	return values, nil
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
