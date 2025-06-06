package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents a MongoDB database connection
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	driverName string
}

// NewMongoDB creates a new MongoDB database connection
func NewMongoDatabase(config *types.Config) (*MongoDB, error) {
	var uri string
	host := config.Host
	if host == "localhost" {
		host = "127.0.0.1"
	}

	if config.Username != "" && config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			config.Username,
			config.Password,
			host,
			config.Port,
			config.Database,
		)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s",
			host,
			config.Port,
			config.Database,
		)
	}

	// Append SSL options to the URI if specified
	if config.SSLMode == "disable" {
		uri += "?ssl=false"
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoDB{
		client:     client,
		database:   client.Database(config.Database),
		driverName: "mongodb",
	}, nil
}

// Connect establishes a connection to the MongoDB database
func (m *MongoDB) Connect(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// Close closes the MongoDB database connection
func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.Background())
}

// CreateTable creates a table in the database
func (m *MongoDB) CreateTable(ctx context.Context, tableName string, table *types.Table, relations []types.Relationship) error {
	// MongoDB doesn't require explicit table creation
	// Just create an index on _id if it's a primary key
	for _, col := range table.Columns {
		if col.IsPrimary {
			_, err := m.database.Collection(tableName).Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys:    bson.D{{Key: col.Name, Value: 1}},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				return fmt.Errorf("failed to create index on %s: %w", col.Name, err)
			}
		}
	}
	return nil
}

// DropTable drops a collection in MongoDB
func (m *MongoDB) DropTable(ctx context.Context, tableName string) error {
	return m.database.Collection(tableName).Drop(ctx)
}

// InsertData inserts data into a MongoDB collection
func (m *MongoDB) InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Convert []map[string]interface{} to []interface{} for MongoDB
	documents := make([]interface{}, len(data))
	for i, doc := range data {
		documents[i] = doc
	}

	_, err := m.database.Collection(tableName).InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to insert data into %s: %w", tableName, err)
	}

	return nil
}

// GetDriver returns the MongoDB driver name
func (m *MongoDB) GetDriver() string {
	return m.driverName
}

// GetAllIDs retrieves all _id values from a collection
func (m *MongoDB) GetAllIDs(ctx context.Context, tableName string) ([]string, error) {
	collection := m.database.Collection(tableName)
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to get IDs from %s: %w", tableName, err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// Log the error but don't return it since this is in a defer
			fmt.Printf("Error closing cursor: %v\n", err)
		}
	}()

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results from %s: %w", tableName, err)
	}

	ids := make([]string, len(results))
	for i, result := range results {
		if id, ok := result["_id"].(string); ok {
			ids[i] = id
		}
	}

	return ids, nil
}

// GetAllForeignKeys retrieves all foreign key values from a collection
func (m *MongoDB) GetAllForeignKeys(ctx context.Context, tableName, columnName string) ([]string, error) {
	collection := m.database.Collection(tableName)
	opts := options.Find().SetProjection(bson.D{{Key: columnName, Value: 1}})
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys from %s.%s: %w", tableName, columnName, err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			// Log the error but don't return it since this is in a defer
			fmt.Printf("Error closing cursor: %v\n", err)
		}
	}()

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results from %s.%s: %w", tableName, columnName, err)
	}

	foreignKeys := make([]string, 0, len(results))
	for _, result := range results {
		if fk, ok := result[columnName].(string); ok {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, nil
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
			}
		}
	}

	// Create the index
	_, err := m.database.Collection(tableName).Indexes().CreateOne(ctx, indexModel)
	return err
}

// VerifyReferentialIntegrity checks if all foreign key references are valid
func (m *MongoDB) VerifyReferentialIntegrity(ctx context.Context, fromTable, fromColumn, toTable, toColumn string) error {
	// Get all IDs from the parent table
	parentIDs, err := m.GetAllIDs(ctx, fromTable)
	if err != nil {
		return fmt.Errorf("failed to get parent IDs: %w", err)
	}

	// Create a map for O(1) lookup
	parentIDMap := make(map[string]bool)
	for _, id := range parentIDs {
		parentIDMap[id] = true
	}

	// Get all foreign key values from the child table
	childFKs, err := m.GetAllForeignKeys(ctx, toTable, toColumn)
	if err != nil {
		return fmt.Errorf("failed to get child foreign keys: %w", err)
	}

	// Check each foreign key value
	var invalidRefs []string
	for _, fk := range childFKs {
		if !parentIDMap[fk] {
			invalidRefs = append(invalidRefs, fk)
		}
	}

	// Report any invalid references
	if len(invalidRefs) > 0 {
		return fmt.Errorf("found %d invalid references in %s.%s: %v",
			len(invalidRefs), toTable, toColumn, invalidRefs)
	}

	return nil
}

// BeginTransaction starts a new MongoDB transaction
func (m *MongoDB) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	session, err := m.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	return &MongoDBTransaction{
		session: session,
		ctx:     ctx,
	}, nil
}

// MongoDBTransaction represents a MongoDB transaction
type MongoDBTransaction struct {
	session mongo.Session
	ctx     context.Context
}

// Commit commits the transaction
func (t *MongoDBTransaction) Commit() error {
	return t.session.CommitTransaction(t.ctx)
}

// Rollback rolls back the transaction
func (t *MongoDBTransaction) Rollback() error {
	return t.session.AbortTransaction(t.ctx)
}

// CreateConstraint creates a constraint on a MongoDB collection
func (m *MongoDB) CreateConstraint(ctx context.Context, tableName string, constraint types.Constraint) error {
	// For MongoDB, we'll create an index on the constraint columns
	// This helps with query performance and ensures uniqueness if needed
	keys := make(bson.D, len(constraint.Columns))
	for i, col := range constraint.Columns {
		keys[i] = bson.E{Key: col, Value: 1}
	}

	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	// Set index options based on constraint type
	switch constraint.Type {
	case "unique":
		indexModel.Options = options.Index().SetUnique(true)
	case "foreign_key":
		// For foreign keys, we'll create a non-unique index
		indexModel.Options = options.Index()
	}

	// Create the index
	_, err := m.database.Collection(tableName).Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create constraint: %w", err)
	}

	return nil
}

// UpdateData updates existing data in a collection
func (m *MongoDB) UpdateData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	collection := m.database.Collection(tableName)

	// Create a bulk write model for updates
	var operations []mongo.WriteModel
	for _, doc := range data {
		// Extract the _id field
		id, ok := doc["_id"]
		if !ok {
			return fmt.Errorf("document missing _id field")
		}

		// Create update operation
		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id}).
			SetUpdate(bson.M{"$set": doc})

		operations = append(operations, update)
	}

	// Execute bulk write
	if len(operations) > 0 {
		opts := options.BulkWrite().SetOrdered(true)
		_, err := collection.BulkWrite(ctx, operations, opts)
		if err != nil {
			return fmt.Errorf("failed to update data in collection %s: %w", tableName, err)
		}
	}

	return nil
}
