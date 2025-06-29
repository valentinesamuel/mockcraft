package mongodb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/term"
)

// MongoDB represents a MongoDB database connection
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	driverName string
	config     *types.Config
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
		config:     config,
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
	// MongoDB automatically creates a unique index on _id, so we don't need to create one
	// Just verify the collection exists by ensuring it has the necessary structure
	
	// For MongoDB, we don't need to create indexes on _id as it's automatically indexed
	// The _id field in MongoDB is automatically unique and indexed by default
	// Creating an explicit unique index on _id would cause an error
	
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
		// Map 'id' field to '_id' for MongoDB primary key
		if idValue, ok := doc["id"]; ok {
			// Ensure _id is of an appropriate type for MongoDB (e.g., string or ObjectID)
			// For now, assuming string, but might need more sophisticated handling later
			doc["_id"] = fmt.Sprintf("%v", idValue)
			delete(doc, "id") // Remove the original 'id' field
		}
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

	// Check if any of the columns is _id
	for _, col := range index.Columns {
		if col == "_id" {
			// MongoDB automatically creates a unique index on _id
			// Cannot create additional indexes on _id field with unique specification
			if index.IsUnique {
				// Skip creating unique index on _id as it's automatically unique
				return nil
			}
		}
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
	// Check if any of the columns is _id and constraint is unique
	for _, col := range constraint.Columns {
		if col == "_id" && constraint.Type == "unique" {
			// MongoDB automatically creates a unique index on _id
			// Skip creating unique constraint on _id as it's automatically unique
			return nil
		}
	}

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

// Backup creates a backup of the database using mongodump --archive
func (m *MongoDB) Backup(ctx context.Context, backupPath string) error {
	log.Printf("Creating backup of database '%s' to '%s' using mongodump --archive", m.database.Name(), backupPath)

	// Check if mongodump is installed and in PATH
	mongodumpPath, err := exec.LookPath("mongodump")
	if err != nil {
		return fmt.Errorf("mongodump is not installed or not in PATH. Please install MongoDB Database Tools: https://www.mongodb.com/try/download/database-tools")
	}

	// Prompt for password
	fmt.Print("Enter MongoDB password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	// Construct the mongodump command
	args := []string{
		fmt.Sprintf("--uri=mongodb://%s@%s:%d", m.config.Username, m.config.Host, m.config.Port),
		fmt.Sprintf("--db=%s", m.database.Name()),
		fmt.Sprintf("--archive=%s", backupPath),
		"--gzip", // Optional: compress the archive
	}

	cmd := exec.CommandContext(ctx, mongodumpPath, args...)

	// Provide password to stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	go func() {
		defer stdin.Close()
		if _, err := io.WriteString(stdin, string(password)+"\n"); err != nil {
			log.Printf("Warning: failed to write password to stdin: %v", err)
		}
	}()

	// Capture stderr for error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		// Include stderr output in the error message
		return fmt.Errorf("mongodump failed: %w\nStderr: %s", err, stderr.String())
	}

	log.Printf("Backup created successfully at '%s'", backupPath)
	return nil
}

// Restore implements types.Database.
func (m *MongoDB) Restore(ctx context.Context, backupFile string) error {
	// Programmatically restore data from the backup directory (backupFile)
	// backupFile in this context is expected to be the directory containing the backup.

	// Check if the backup directory exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup directory %s does not exist: %w", backupFile, err)
	}

	// Connect to MongoDB
	host := m.config.Host
	if host == "localhost" {
		host = "127.0.0.1"
	}

	var uri string
	if m.config.Username != "" && m.config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			m.config.Username,
			m.config.Password,
			host,
			m.config.Port,
			m.config.Database,
		)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s",
			host,
			m.config.Port,
			m.config.Database,
		)
	}

	// Append SSL options to the URI if specified
	if m.config.SSLMode == "disable" {
		uri += "?ssl=false"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Ping the primary to ensure the connection is established
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(m.config.Database)

	// Read the contents of the backup directory
	files, err := os.ReadDir(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup directory %s: %w", backupFile, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories if any
		}

		// Assume each file is a collection name.json
		filename := file.Name()
		if !strings.HasSuffix(filename, ".json") {
			continue // Skip non-json files
		}

		collectionName := strings.TrimSuffix(filename, ".json")
		collection := db.Collection(collectionName)

		filePath := filepath.Join(backupFile, filename)

		// Read the JSON file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read backup file %s: %w", filePath, err)
		}

		// Unmarshal the JSON data. Assuming the backup is an array of documents.
		var documents []bson.M
		err = bson.UnmarshalExtJSON(data, false, &documents)
		if err != nil {
			// Handle potential errors if the file is not a JSON array of objects
			// For simplicity, log the error and skip the file.
			log.Printf("Skipping file %s due to JSON unmarshalling error: %v", filePath, err)
			continue
		}

		// If there are documents, insert them into the collection
		if len(documents) > 0 {
			// Insert many documents
			insertManyOptions := options.InsertMany().SetOrdered(false) // Continue on errors
			insertResult, err := collection.InsertMany(ctx, bsonMDocumentsToInterfaceSlice(documents), insertManyOptions)
			if err != nil {
				// Log the error but attempt to continue with other collections
				log.Printf("Failed to insert documents into collection %s from file %s: %v", collectionName, filePath, err)
				// Depending on requirements, you might want to return the error here.
				// For now, we log and continue.
			} else {
				log.Printf("Successfully inserted %d documents into collection %s", len(insertResult.InsertedIDs), collectionName)
			}
		} else {
			log.Printf("No documents to insert for collection %s from file %s", collectionName, filePath)
		}
	}

	return nil
}

// Helper function to convert []bson.M to []interface{}
func bsonMDocumentsToInterfaceSlice(docs []bson.M) []interface{} {
	interfaces := make([]interface{}, len(docs))
	for i, doc := range docs {
		interfaces[i] = doc
	}
	return interfaces
}
