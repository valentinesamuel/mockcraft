# MockCraft - Universal Fake Data Generator

A comprehensive fake data toolkit with three modes: CLI faker, database seeder, and REST API server.

## 📋 Implementation Requirements

### Core Features
- **Three-in-one tool**: CLI faker, database seeder, REST API server
- **Single binary** with subcommands for all operating systems
- **Industry-specific generators** extending gofakeit (health, aviation, finance, etc.)
- **Multiple database support**: Both relational (PostgreSQL, MySQL, SQLite) and NoSQL (MongoDB, Redis)
- **Multiple output formats**: CSV, JSON, SQL dumps
- **Database backup functionality**

### CLI Requirements

#### 1. Faker CLI
```bash
mockcraft generate firstname                    # → Isabella
mockcraft generate password --length=12        # → asne45p0gjnw56ghw
mockcraft generate medical_condition           # → Hypertension
```

#### 2. Database Seeder CLI
```bash
mockcraft seed --config schema.yaml --db postgres://...
mockcraft seed --config schema.yaml --output csv --dir ./output
mockcraft seed --config schema.yaml --backup --backup-path ./backup/
```

#### 3. Server CLI
```bash
mockcraft server --port 8080 --config server.yaml
```

### REST API Requirements

#### Core Endpoints
- `GET /api/generate/{type}` - Generate single fake data
- `GET /api/generators` - List all available generators
- `POST /api/seed` - Upload YAML, get job ID for async processing
- `GET /api/jobs/{id}` - Check job status
- `GET /api/download/{id}` - Download generated files (ZIP)
- `GET /metrics` - Prometheus-style metrics

#### Server Features
- **Per-IP rate limiting** (configurable limits)
- **Async job processing** for large file generation
- **Request logging** with structured logs
- **CORS support** for frontend integration
- **File size limits** for uploads
- **Job cleanup** after download/expiry

### Configuration Requirements
- **CLI flags** for all options
- **Config file support** (YAML format)
- **Environment variable** override support

### Build Requirements
- **Cross-compilation** for Linux (amd64, arm64), Windows (amd64), macOS (amd64, arm64)
- **Single binary** distribution
- **Build automation** scripts

---

## 🗺️ Step-by-Step Implementation Roadmap

### Phase 1: Project Foundation
**Goal**: Set up project structure and basic CLI framework

#### Step 1.1: Initialize Project Structure
```bash
mkdir mockcraft && cd mockcraft
go mod init github.com/yourusername/mockcraft
```

Create the complete folder structure:
```
mockcraft/
├── cmd/mockcraft/
├── internal/{config,generators,database,output,server,backup}/
├── pkg/utils/
├── configs/
├── scripts/
└── examples/
```

#### Step 1.2: Basic CLI Framework
- Implement root command with Cobra CLI
- Add subcommands: `generate`, `seed`, `server`
- Basic flag parsing and validation
- Help system and version command

**Deliverable**: `mockcraft --help` works with all subcommands listed

---

### Phase 2: Core Generator System
**Goal**: Implement the fake data generation engine

#### Step 2.1: Base Generator Registry
- Create generator registry pattern
- Wrap gofakeit with custom interface
- Implement basic type mapping
- Add parameter parsing system

#### Step 2.2: Industry-Specific Generators
- Health/Medical generators (blood types, conditions, medications)
- Aviation generators (airlines, airports, flight numbers)
- Finance generators (account numbers, routing numbers, currencies)
- Extensible plugin system for future additions

#### Step 2.3: CLI Generate Command
- Implement `mockcraft generate <type>` functionality
- Add parameter support (--length, --format, etc.)
- Error handling and validation
- Output formatting options

**Deliverable**: `mockcraft generate firstname` returns fake names

---

### Phase 3: Configuration and Schema System
**Goal**: YAML schema parsing and configuration management

#### Step 3.1: Configuration Management
- YAML config file parsing
- Environment variable support
- CLI flag precedence system
- Validation and defaults

#### Step 3.2: Schema Definition System
- YAML schema parser for database seeding
- Table relationship handling
- Foreign key dependency resolution
- Data type validation and constraints

#### Step 3.3: Schema Validation
- Validate schema syntax
- Check generator availability
- Dependency cycle detection
- Error reporting with line numbers

**Deliverable**: Parse and validate complex YAML schemas

---

### Phase 4: Database Integration
**Goal**: Multi-database support with backup functionality

#### Step 4.1: Database Abstraction Layer
- Common interface for all database types
- Connection management and pooling
- Transaction handling
- Error standardization

#### Step 4.2: Relational Database Support
- PostgreSQL implementation
- MySQL implementation  
- SQLite implementation
- Bulk insert optimization
- Schema introspection

#### Step 4.3: NoSQL Database Support
- MongoDB implementation
- Redis implementation
- Document/key-value abstractions
- Batch operations

#### Step 4.4: Backup System
- SQL dump generation (pg_dump, mysqldump)
- NoSQL export functionality
- Compression and archiving
- Restore verification

**Deliverable**: Connect to multiple database types and perform backups

---

### Phase 5: Output Generation
**Goal**: Multiple output format support

#### Step 5.1: CSV Output
- Efficient CSV generation
- Proper escaping and encoding
- Large file streaming
- Memory optimization

#### Step 5.2: JSON Output
- Structured JSON generation
- Nested object support
- Array handling for relationships
- Pretty printing options

#### Step 5.3: SQL Dump Output
- INSERT statement generation
- Database-specific syntax
- Transaction wrapping
- Constraint handling

**Deliverable**: Generate data in multiple formats from same schema

---

### Phase 6: Database Seeder CLI
**Goal**: Complete CLI seeding functionality

#### Step 6.1: Seed Command Implementation
- Schema loading and parsing
- Data generation pipeline
- Progress reporting
- Error recovery

#### Step 6.2: Output Options
- File output (CSV, JSON, SQL)
- Direct database insertion
- Batch processing control
- Memory management

#### Step 6.3: Advanced Features
- Incremental seeding
- Data relationships
- Custom constraints
- Performance optimization

**Deliverable**: `mockcraft seed` fully functional with all database types

---

### Phase 7: REST API Server
**Goal**: HTTP server with async job processing

#### Step 7.1: Basic HTTP Server
- Gin/Echo framework setup
- Route definitions
- Middleware pipeline
- Error handling

#### Step 7.2: Generator Endpoints
- `/api/generate/{type}` implementation
- Parameter parsing from query strings
- Response formatting (JSON)
- Input validation

#### Step 7.3: Generator Discovery
- `/api/generators` endpoint
- Dynamic generator listing
- Parameter documentation
- Category grouping

**Deliverable**: Basic API server responding to generate requests

---

### Phase 8: Async Job Processing
**Goal**: Handle large file generation asynchronously

#### Step 8.1: Job Queue System
- In-memory job queue
- Job status tracking
- Worker pool management
- Progress reporting

#### Step 8.2: File Processing Pipeline
- YAML upload handling
- Validation and parsing
- Background generation
- File compression (ZIP)

#### Step 8.3: Job Management API
- `/api/seed` - Create jobs
- `/api/jobs/{id}` - Status checking
- `/api/download/{id}` - File download
- Job cleanup and expiry

**Deliverable**: Upload YAML, get job ID, download ZIP of results

---

### Phase 9: Middleware and Security
**Goal**: Production-ready server features

#### Step 9.1: Rate Limiting
- Per-IP rate limiting
- Configurable limits
- Different limits per endpoint
- Redis backing (optional)

#### Step 9.2: Logging and Monitoring
- Structured request logging
- Error tracking
- Performance metrics
- `/metrics` Prometheus endpoint

#### Step 9.3: CORS and Security
- CORS middleware
- File upload size limits
- Input sanitization
- Security headers

**Deliverable**: Production-ready server with monitoring

---

### Phase 10: Build and Distribution
**Goal**: Cross-platform binary distribution

#### Step 10.1: Build System
- Makefile for common tasks
- Cross-compilation scripts
- Version management
- Binary optimization

#### Step 10.2: CI/CD Pipeline
- GitHub Actions setup
- Automated testing
- Multi-platform builds
- Release automation

#### Step 10.3: Documentation
- Complete API documentation
- Usage examples
- Configuration reference
- Troubleshooting guide

**Deliverable**: Distributable binaries for all platforms

---

## 🎯 Milestone Checklist

- [ ] **Phase 1**: Basic CLI structure working
- [ ] **Phase 2**: `mockcraft generate` command functional
- [ ] **Phase 3**: YAML schema parsing complete
- [ ] **Phase 4**: Database connections working
- [ ] **Phase 5**: Multiple output formats supported
- [ ] **Phase 6**: `mockcraft seed` command complete
- [ ] **Phase 7**: Basic REST API server running
- [ ] **Phase 8**: Async job processing working
- [ ] **Phase 9**: Production middleware complete
- [ ] **Phase 10**: Cross-platform builds ready

## 📚 Key Dependencies

```go
// Core dependencies
"github.com/spf13/cobra"           // CLI framework
"github.com/brianvoe/gofakeit/v6"  // Fake data generation
"github.com/gin-gonic/gin"         // HTTP framework
"gopkg.in/yaml.v3"                 // YAML parsing

// Database drivers
"github.com/lib/pq"                // PostgreSQL
"github.com/go-sql-driver/mysql"   // MySQL
"github.com/mattn/go-sqlite3"      // SQLite
"go.mongodb.org/mongo-driver"      // MongoDB
"github.com/go-redis/redis/v8"     // Redis

// Utilities
"github.com/schemalex/difflib"     // Progress bars
"golang.org/x/time/rate"           // Rate limiting
"github.com/prometheus/client_golang" // Metrics
```

---

Each phase builds upon the previous one, ensuring a solid foundation before adding complexity. Start with Phase 1 and work sequentially through each step.
