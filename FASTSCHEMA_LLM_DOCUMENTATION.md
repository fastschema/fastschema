# FastSchema - Comprehensive LLM Documentation

## Project Overview

FastSchema is a Backend as a Service (BaaS) and Go web framework for building dynamic web applications. It automates database generation, provides ready-to-use CRUD APIs, and offers tools for managing structured content effortlessly.

**Key Information:**
- **Language:** Go (version 1.24+)
- **License:** MIT
- **Repository:** https://github.com/fastschema/fastschema
- **Documentation:** https://fastschema.com/docs/
- **Current Version:** 0.6.2+
- **Author:** Nguyen Ngoc Phuong and Contributors

## Core Architecture

### Main Components

1. **App Structure** (`fastschema.go`)
   - Central application struct managing all components
   - Handles configuration, database, resources, and services
   - Provides hooks system for extensibility

2. **Schema Builder** (`schema/builder.go`)
   - Manages database schema definitions
   - Handles relationships between schemas
   - Supports JSON-based schema definitions

3. **Database Layer** (`db/db.go`)
   - Supports MySQL, PostgreSQL, SQLite
   - ORM-like interface with query builders
   - Migration system with Atlas integration

4. **Resource System** (`fs/`)
   - RESTful API endpoint management
   - Request/response handling
   - Middleware support

## Project Structure

```
fastschema/
├── cmd/                    # CLI application entry point
├── dash/                   # Admin dashboard (Next.js)
├── db/                     # Database layer and ORM
├── entity/                 # Entity definitions
├── examples/               # Example implementations
├── expr/                   # Expression evaluation
├── fs/                     # Core framework interfaces
├── logger/                 # Logging utilities
├── pkg/                    # Packages and utilities
│   ├── auth/              # Authentication providers
│   ├── dash/              # Dashboard source code
│   ├── entdbadapter/      # Ent ORM adapter
│   └── errors/            # Error handling
├── plugins/               # Plugin system
├── schema/                # Schema management
├── services/              # Core services
│   ├── auth/             # Authentication service
│   ├── content/          # Content management
│   ├── file/             # File management
│   └── realtime/         # Real-time updates
└── tests/                 # Test suites
```

## Key Features

### 1. Automated Database Generation
- Automatically generates database tables from schema definitions
- Flexible relationship models (O2O, O2M, M2M)
- Support for indexes and constraints

### 2. RESTful API Generation
- Auto-generated CRUD endpoints
- OpenAPI specification generation
- Real-time API updates when schemas change

### 3. Dynamic Content Modeling
- JSON-based schema definitions
- Admin UI for schema management
- Instant schema updates

### 4. Built-in Features
- File management system
- Admin control panel
- Role-based access control
- Real-time updates via WebSocket
- Plugin system

### 5. Database Support
- **MySQL:** Full support with migrations
- **PostgreSQL:** Complete compatibility
- **SQLite:** Default database for development

## Installation Methods

### Method 1: Docker
```bash
docker pull ghcr.io/fastschema/fastschema:latest
mkdir data
docker run \
  -u "$UID" \
  -p 8000:8000 \
  -v ./data:/fastschema/data \
  ghcr.io/fastschema/fastschema:latest
```

### Method 2: Binary Download
1. Download from GitHub Releases
2. Extract the binary
3. Run: `./fastschema start`

### Method 3: Build from Source
```bash
git clone https://github.com/fastschema/fastschema.git
cd fastschema
git submodule update --init --recursive
go build -o fastschema cmd/main.go
./fastschema start
```

## Configuration

FastSchema uses environment variables for configuration, stored in `./data/.env`:

### Essential Variables
- `APP_KEY`: 32-character encryption key (required)
- `APP_PORT`: Server port (default: 8000)
- `APP_BASE_URL`: Base application URL
- `APP_DASH_URL`: Dashboard URL
- `DB_DRIVER`: Database driver (sqlite/mysql/pgx)
- `DB_NAME`: Database name
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASS`: Database password

### Advanced Configuration
- `AUTH`: JSON string for authentication providers
- `STORAGE`: JSON string for file storage configuration
- `MAIL`: JSON string for email configuration

## Core Concepts

### Schema
A blueprint defining data structure with:
- **name**: Unique identifier
- **namespace**: Database table/API endpoint name
- **label_field**: Display field for items
- **fields**: Array of field definitions
- **db**: Database configuration (indexes, constraints)

### Field Types
- **Primitives**: bool, string, text, int variants, uint variants, float32/64
- **Special**: time, json, uuid, bytes, enum
- **Relations**: relation, file (special relation to file schema)

### Field Properties
- **name**: Field identifier
- **type**: Data type
- **label**: Human-readable name
- **optional**: Nullable field
- **unique**: Unique constraint
- **size**: Maximum length
- **default**: Default value
- **db**: Database-specific configuration
- **relation**: Relationship configuration

### Relationship Types
- **O2O (One-to-One)**: Single item relationship
- **O2M (One-to-Many)**: Parent-children relationship
- **M2M (Many-to-Many)**: Junction table relationship

### Resource
Components that process requests and return responses:
- Similar to MVC controllers
- Can be grouped for organization
- Must have handler functions
- Support input/output validation
- Routed by RestfulResolver

## Directory Structure (Runtime)

```
data/
├── .env                    # Environment configuration
├── fastschema.db          # SQLite database (default)
├── logs/
│   └── app.log           # Application logs
├── migrations/           # Database migration files
│   ├── *.up.sql         # Upgrade scripts
│   ├── *.down.sql       # Downgrade scripts
│   └── atlas.sum        # Migration checksums
├── public/              # File uploads
└── schema/              # Schema JSON files
```

## API Architecture

### Automatic Endpoints
For each schema, FastSchema generates:
- `GET /api/{schema}` - List records
- `POST /api/{schema}` - Create record
- `GET /api/{schema}/{id}` - Get single record
- `PUT /api/{schema}/{id}` - Update record
- `DELETE /api/{schema}/{id}` - Delete record
- `PATCH /api/{schema}` - Bulk update
- `DELETE /api/{schema}` - Bulk delete

### Authentication Endpoints
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `POST /api/auth/register` - User registration
- `GET /api/auth/me` - Current user info

### File Management
- `POST /api/file/upload` - File upload
- `GET /api/file/{id}` - File download
- `DELETE /api/file/{id}` - File deletion

## Framework Usage

### Basic Application
```go
package main

import (
    "github.com/fastschema/fastschema"
    "github.com/fastschema/fastschema/fs"
)

func main() {
    app, _ := fastschema.New(&fs.Config{
        SystemSchemas: []any{Tag{}, Blog{}},
    })
    
    // Add custom resource
    app.API().Add(fs.Post("/custom", handler))
    
    app.Start()
}
```

### Custom Resource Handler
```go
func handler(c fs.Context, input *InputType) (*OutputType, error) {
    // Process request
    return &OutputType{}, nil
}
```

### Database Operations
```go
// Query
entities, err := db.Builder[*Entity](app.DB()).
    Where(db.EQ("status", "active")).
    Get(ctx)

// Create
id, err := db.Mutation[Entity](app.DB()).
    Create(ctx, entity)

// Update
affected, err := db.Mutation[Entity](app.DB()).
    Where(db.EQ("id", 1)).
    Update(ctx, updateData)
```

## Hooks System

### Application Hooks
- `OnPreResolve`: Before request processing
- `OnPostResolve`: After request processing

### Database Hooks
- `OnPreDBQuery/OnPostDBQuery`: Query operations
- `OnPreDBCreate/OnPostDBCreate`: Create operations
- `OnPreDBUpdate/OnPostDBUpdate`: Update operations
- `OnPreDBDelete/OnPostDBDelete`: Delete operations
- `OnPreDBExec/OnPostDBExec`: Raw SQL execution

## Plugin System

Plugins extend FastSchema functionality:
- JavaScript-based plugin system
- Configuration via JSON
- API access for custom endpoints
- Rule-based access control

## FastSchema Web Framework

FastSchema is a flexible and extensible application that allows you to customize and extend its functionality to meet your specific requirements. It's built using Go and leverages several open-source libraries:

### Core Dependencies
- **[Fiber](https://gofiber.io/)**: High-performance HTTP framework for Go
- **[Ent](https://entgo.io/)**: Entity framework for Go with code generation
- **[Rclone](https://rclone.org/)**: Cloud storage abstraction layer
- **[Zap](https://pkg.go.dev/go.uber.org/zap)**: Structured, leveled logging
- **[Next.js](https://nextjs.org/)**: React framework for the dashboard
- **[Shadcn](https://ui.shadcn.com/)**: Modern UI component library
- **[TipTap](https://www.tiptap.dev/)**: Headless rich text editor

### Prerequisites
- Go 1.18 or later
- Code editor (Visual Studio Code, GoLand, etc.)
- Command-line terminal
- Basic understanding of Go programming

### Installation
```bash
go get github.com/fastschema/fastschema
```

### Basic Usage
```go
package main

import (
    "github.com/fastschema/fastschema/fs"
    "github.com/fastschema/fastschema"
)

func main() {
    app, _ := fastschema.New(&fs.Config{
        Port: "8000",
    })
    
    app.AddResource(fs.Get("/", func(c fs.Context, _ any) (string, error) {
        return "Hello World", nil
    }))
    
    app.Start()
}
```

### Framework Examples

#### Create Public Resources
By default, all resources require authentication. To create public resources:
```go
app.AddResource(fs.Get("/", func(c fs.Context, _ any) (string, error) {
    return "Hello World", nil
}, &fs.Meta{Public: true}))
```

#### HTML Response
```go
app.AddResource(fs.Get("/about", func(c fs.Context, _ any) (any, error) {
    header := make(http.Header)
    header.Set("Content-Type", "text/html")
    return &fs.HTTPResponse{
        StatusCode: http.StatusOK,
        Header:     header,
        Body:       []byte(`<!DOCTYPE html><html>
            <head><title>About</title></head>
            <body><h1>About</h1></body>
        </html>`),
    }, nil
}, &fs.Meta{Public: true}))
```

#### API Group Resources
```go
app.API().Add(fs.Get("/hello", func(c fs.Context, _ any) (any, error) {
    return fs.Map{"message": "Hello, World!"}, nil
}))
```

#### Resource Groups
```go
app.Resources().
    Group("docs", &fs.Meta{Prefix: "/docs"}).
    Add(fs.Get("/getting-started", func(c fs.Context, _ any) (any, error) {
        return fs.Map{"message": "Getting started with FastSchema"}, nil
    }, &fs.Meta{Public: true}))
```

## Development Commands

```bash
# Development server with hot reload
make dev

# Run tests
go test ./...

# Integration tests (requires Docker)
cd tests/integration
docker compose up -d
go test ./...

# Build binary
go build -o fastschema cmd/main.go

# Setup application
./fastschema setup -u admin -p password

# Reset admin password
./fastschema reset-admin-password -p newpassword
```

## Use Cases

### 1. Backend as a Service (No-Code)
- Headless CMS capabilities
- API-first development
- Dynamic content modeling
- Real-time data management
- Zero code required for basic functionality

### 2. Web Framework
- Custom endpoint creation via Resources
- Extensible through Hooks
- Powerful ORM for database operations
- Built-in authentication and authorization
- File management and storage

## Security Features

### Authentication
- JWT-based authentication
- Multiple auth providers (local, GitHub, Google, Twitter)
- Password reset functionality
- Account activation via email

### Authorization
- Role-based access control (RBAC)
- Permission system
- Resource-level security
- Field-level access control

### Data Protection
- SQL injection prevention
- Input validation
- CORS support
- Secure file uploads

## Performance Features

- Connection pooling
- Query optimization
- Caching support
- Real-time updates via WebSocket
- Efficient file serving
- Database indexing

## Deployment

FastSchema can be deployed as:
- Docker container
- Standalone binary
- Embedded in Go applications
- Cloud platforms (AWS, GCP, Azure)
- Traditional servers

## Monitoring and Logging

- Structured logging with Zap
- Request/response logging
- Database query logging
- Error tracking
- Performance metrics
- Health check endpoints

## Plugin System Deep Dive

### Overview

The FastSchema Plugins System allows developers to extend core functionality by writing JavaScript code, eliminating the need for deep Golang familiarity. This empowers users to introduce custom features, manage content, and interact with FastSchema's core APIs through a flexible plugin architecture.

### JavaScript Runtime

FastSchema uses **[Goja](https://github.com/dop251/goja)** as the JavaScript runtime - an ECMAScript 5.1 engine written in pure Go, designed for embedding in Go applications to run scripts and evaluate expressions.

### Plugin Capabilities

Plugins can be used to:
- Add custom schemas
- Add custom resources (API endpoints)
- Add custom hooks (lifecycle events)
- Modify FastSchema configuration
- Add custom logic to lifecycle events

### JavaScript Plugin Architecture

#### Plugin Structure

A FastSchema plugin is a directory located in `data/plugins` containing a `plugin.js` file as the entry point:

```
fastschema-app
└── data
    └── plugins
        └── hello
            ├── plugin.js          # Main plugin file with Config and Init functions
            ├── resources.js       # Custom API endpoints
            ├── hooks.js          # Database and application hooks
            ├── utils.js          # Utility functions
            ├── product.json      # Schema definitions
            └── schemas/          # Additional schema files
                └── schema.json
```

#### Core Plugin Files

**plugin.js** - Main entry point:
```javascript
const product = require('./schemas/product.json');
const { getRandomName, ping } = require('./utils');

/** @param {FsAppConfigActions} config */
const Config = config => {
  // Add product schema
  config.AddSchemas(product);
  
  // Change the fastschema port to 9000
  config.port = '9000';
}

/** @param {FsPlugin} plugin */
const Init = plugin => {
  // Create a group named 'hello' with two public resources (routes):
  // - hello/ping
  // - hello/world
  plugin.resources
    .Group('hello')
    .Add(ping, { public: true })
    .Add(world, { public: true });
}

const world = async ctx => {
  const name = await getRandomName();
  return `Hello, ${name}!`;
}
```

**utils.js** - Utility functions:
```javascript
const getRandomName = async () => {
  const names = ['Alice', 'Bob', 'Charlie', 'David', 'Eve', 'Frank', 'Grace', 'Hank', 'Ivy', 'Jack'];
  return names[Math.floor(Math.random() * names.length)];
}

const ping = ctx => {
  return 'pong';
}

module.exports = {
  getRandomName,
  ping,
}
```

**schemas/product.json** - Schema definition:
```json
{
  "name": "product",
  "namespace": "products",
  "label_field": "name",
  "fields": [
    {
      "type": "string",
      "name": "name",
      "label": "Name",
      "optional": true,
      "sortable": true
    },
    {
      "type": "string",
      "name": "description",
      "label": "Description",
      "optional": true,
      "sortable": true,
      "renderer": {
        "class": "w-full bg-gray-50 rounded-lg border"
      }
    }
  ]
}
```

**resources.js** - API endpoints with database operations:
```javascript
/** @param {FsContext} ctx */
const ping = ctx => {
  return 'pong';
}

/** @param {FsContext} ctx */
const world = async ctx => {
  const tx = $db().Tx(ctx);
  const roles = tx.Query(ctx, 'SELECT * FROM roles WHERE id IN ($1, $2)', 1, 3);
  return { data: 'Hello, World', roles };
}

/** @param {FsContext} ctx */
const createProduct = async ctx => {
  const tx = $db().Tx(ctx);
  try {
    const product = await tx.Create(ctx, 'product', {
      name: ctx.Arg('name'),
      description: ctx.Arg('description')
    });
    await tx.Commit(ctx);
    return { success: true, product };
  } catch (error) {
    await tx.Rollback(ctx);
    throw error;
  }
}

/** @param {FsContext} ctx */
const getProducts = async ctx => {
  const products = await $db().Query(ctx, 'SELECT * FROM products ORDER BY created_at DESC');
  return { products };
}
```

#### Plugin Lifecycle

The FastSchema Plugins System follows a three-stage lifecycle:

1. **Configuration**: Called right after FastSchema starts bootstrapping
   - Only has access to FastSchema configuration object
   - Can modify configuration as needed
   - Example: Change port, add schemas

2. **Initialization**: Called after FastSchema application is bootstrapped, before server starts
   - Access to FastSchema APIs: logger, db, context, resources
   - Register resources and hooks
   - Set up plugin functionality

3. **Execution**: Plugin resources and hooks execute during application lifecycle
   - Resources handle incoming requests
   - Hooks intercept and modify event flow

#### Plugin Capabilities

1. **Schema Management**: Add custom schemas via `config.AddSchemas()`
2. **Database Access**: Full database operations through `$db()` global
3. **Transaction Support**: `tx = $db().Tx(ctx)` with rollback/commit
4. **Hook System**: All application and database hooks available
5. **Resource Creation**: Custom API endpoints with full HTTP support
6. **Context Access**: Request context, user info, logging, arguments

#### TypeScript Definitions

FastSchema provides comprehensive TypeScript definitions for plugins:

- `FsContext`: Request context interface
- `FsLogger`: Logging interface
- `FsResource`: Resource management
- `FsAppConfig`: Application configuration
- `FsUser`, `FsRole`, `FsFile`: System entities

### Go Plugin Development

#### System Schema Definition
```go
type Blog struct {
    ID    uint64 `json:"id"`
    Title string `json:"title"`
    Body  string `json:"body" fs:"optional;type=text" fs.renderer:"{'class':'editor'}"`
    Vote  int    `json:"vote" fs:"optional"`
    Tags  []*Tag `json:"tags" fs:"optional" fs.relation:"{'type':'m2m','schema':'tag','field':'blogs'}"`
}
```

#### Field Tags System
- `fs:"unique"` - Unique constraint
- `fs:"optional"` - Nullable field
- `fs:"type=text"` - Field type override
- `fs.renderer:"{'class':'editor'}"` - UI renderer configuration
- `fs.relation:"{'type':'m2m','schema':'tag','field':'blogs'}"` - Relationship definition

#### Relationship Types
1. **One-to-One (o2o)**: `{'type':'o2o','schema':'profile','field':'student','owner':true}`
2. **One-to-Many (o2m)**: `{'type':'o2m','schema':'comment','field':'post','owner':true}`
3. **Many-to-Many (m2m)**: `{'type':'m2m','schema':'category','field':'posts'}`

## Dashboard Architecture Analysis

### Technology Stack
- **Framework**: Next.js 14 with TypeScript
- **UI Library**: Radix UI components with Shadcn/ui
- **Styling**: Tailwind CSS with custom design system
- **Rich Text**: TipTap editor with extensive extensions
- **State Management**: TanStack Query for server state
- **Forms**: React Hook Form with Zod validation
- **Code Editor**: CodeMirror with JavaScript/Rust support

### Project Structure
```
pkg/dash/src/
├── app/                    # Next.js app router
│   ├── content/           # Content management pages
│   ├── schemas/           # Schema management
│   ├── settings/          # Application settings
│   └── media/            # File management
├── components/
│   ├── ui/               # Shadcn/ui components
│   ├── common/           # Shared components
│   ├── content/          # Content-specific components
│   ├── schema/           # Schema management components
│   └── media/            # Media management components
├── lib/                   # Utilities and configurations
└── hooks/                # Custom React hooks
```

### Key Features
1. **Dynamic Schema Management**: Real-time schema creation and editing
2. **Rich Content Editor**: TipTap-based WYSIWYG editor
3. **File Management**: Drag-and-drop file uploads with preview
4. **Role-Based UI**: Dynamic interface based on user permissions
5. **Real-time Updates**: WebSocket integration for live data
6. **Responsive Design**: Mobile-first responsive layout
7. **Dark/Light Mode**: Theme switching with system preference detection

### Build Configuration
- **Output**: Static export (`output: 'export'`)
- **Base Path**: `/dash` for integration with Go backend
- **Bundle Analysis**: Webpack bundle analyzer integration
- **TypeScript**: Strict mode with comprehensive type checking

## Examples and Coding Patterns

### 1. Database Operations Example
```go
// Create with relationships
blog1 := utils.Must(db.Builder[Blog](app.DB()).Create(ctx, fs.Map{
    "title": "Blog 1",
    "body":  "Blog 1 body",
    "tags": []*entity.Entity{
        entity.New(tag1.ID),
        tag2,
    },
}))

// Query with relationships
blog1 = utils.Must(db.Builder[Blog](app.DB()).
    Where(db.EQ("id", blog1.ID)).
    Select("tags").
    First(ctx))

// Raw SQL queries
blog1Tags := utils.Must(db.Query[*entity.Entity](
    ctx, app.DB(),
    "SELECT * FROM tags JOIN blogs_tags ON tags.id = blogs_tags.tags WHERE blogs_tags.blogs = ?",
    blog1.ID,
))
```

### 2. Hook System Example
```go
// Application hooks
app.OnPreResolve(func(c fs.Context) error {
    c.Local("custom", "pre resolve hook")
    return nil
})

// Database hooks
app.OnPostDBQuery(func(
    ctx context.Context,
    query *db.QueryOption,
    entities []*entity.Entity,
) ([]*entity.Entity, error) {
    for _, entity := range entities {
        entity.Set("desc", fmt.Sprintf("Description for %s", entity.Get("name")))
    }
    return entities, nil
})
```

### 3. Resource Creation Patterns
```go
// Simple GET endpoint
app.API().Add(fs.Get("/hello", func(c fs.Context, _ any) (any, error) {
    return fs.Map{"message": "Hello, World!"}, nil
}, &fs.Meta{Public: true}))

// POST with typed input/output
app.API().Add(fs.Post("/blog", func(c fs.Context, data *BlogCreate) (*BlogLink, error) {
    return &BlogLink{
        ID:   1,
        Name: data.Name,
        Link: "/blog/" + data.Name,
    }, nil
}))

// HTML response
app.AddResource(fs.Get("/about", func(c fs.Context, _ any) (any, error) {
    header := make(http.Header)
    header.Set("Content-Type", "text/html")
    return &fs.HTTPResponse{
        StatusCode: http.StatusOK,
        Header:     header,
        Body:       []byte(`<html>...</html>`),
    }, nil
}, &fs.Meta{Public: true}))
```

### 4. Storage Management
```go
// Default local storage
disk := app.Disk()
file := utils.Must(disk.Put(ctx, &fs.File{
    Name:   "file.txt",
    Path:   "custom/file.txt",
    Type:   "text/plain",
    Size:   11,
    Reader: bytes.NewReader([]byte("Hello world")),
}))

// Custom storage configuration
app := utils.Must(fastschema.New(&fs.Config{
    StorageConfig: &fs.StorageConfig{
        DefaultDisk: "local_public",
        Disks: []*fs.DiskConfig{
            {
                Name:       "local_public",
                Driver:     "local",
                Root:       "./public",
                BaseURL:    "http://localhost:8000/files",
                PublicPath: "/files",
            },
        },
    },
}))
```

### 5. Relationship Modeling
```go
// One-to-One relationship
type Student struct {
    ID      int      `json:"id"`
    Name    string   `json:"name"`
    Profile *Profile `json:"profile" fs.relation:"{'type':'o2o','schema':'profile','field':'student','owner':true}"`
}

// One-to-Many relationship
type Post struct {
    ID       int        `json:"id"`
    Title    string     `json:"title"`
    Comments []*Comment `json:"comments" fs.relation:"{'type':'o2m','schema':'comment','field':'post','owner':true}"`
}

// Many-to-Many relationship
type Category struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Posts []*Post `json:"posts" fs.relation:"{'type':'m2m','schema':'post','field':'categories'}"`
}
```

## Coding Best Practices

### 1. Error Handling
- Use `utils.Must()` for examples and prototyping
- Implement proper error handling in production code
- Leverage `db.IsNotFound(err)` for database error checking

### 2. Context Management
- Always pass `context.Context` for database operations
- Use `c.Context()` in resource handlers for request context
- Implement proper context cancellation for long-running operations

### 3. Type Safety
- Define strong types for API inputs/outputs
- Use generic database builders: `db.Builder[Type](app.DB())`
- Leverage TypeScript definitions in JavaScript plugins

### 4. Resource Organization
- Group related resources using `app.Resources().Group()`
- Use meaningful resource names and paths
- Implement proper meta information for OpenAPI generation

### 5. Plugin Development
- Separate concerns: resources, hooks, utilities
- Use TypeScript definitions for better development experience
- Implement proper error handling in JavaScript code
- Leverage transaction support for data consistency

## Comprehensive Hooks System

Hooks are functions that execute before or after operations on resources. FastSchema provides three types of hooks: resource hooks, database hooks, and application hooks, allowing custom logic extension.

### Hook Types and Signatures

Most hooks share similar signatures with context and specific parameters:

```go
// Context object passed to actions
context.Context

// Middleware function for resources
type Middleware func(c Context) error

// Query options for database operations
type QueryOption struct {
    Schema     *schema.Schema `json:"schema"`
    Limit      uint          `json:"limit"`
    Offset     uint          `json:"offset"`
    Columns    []string      `json:"columns"`
    Order      []string      `json:"order"`
    Predicates []*Predicate  `json:"predicates"`
    Query      string        `json:"query"`
    Args       any           `json:"args"`
    // For count queries
    Column     string        `json:"column"`
    Unique     bool          `json:"unique"`
}
```

### Application Hooks

#### OnPreResolve
```go
func (a *App) OnPreResolve(middlewares ...fs.Middleware)
```

Executed before resource handler is called. Can add custom logic or stop execution:

- Return `nil`: Continue normal flow
- Return `error`: Stop execution and return error to client

```go
app.OnPreResolve(func(ctx fs.Context) error {
    user := ctx.User()
    resource := ctx.Resource()
    if user == nil && resource.ID() == "api.user.me" {
        return errors.Unauthenticated("You are not authenticated")
    }
    return nil
})
```

#### OnPostResolve
```go
func (a *App) OnPostResolve(middlewares ...fs.Middleware)
```

Executed after resource handler. Can manipulate response before sending to client:

```go
app.OnPostResolve(func(ctx fs.Context) error {
    // Only modify "api.hello" resource response
    if ctx.Resource().ID() != "api.hello" {
        return nil
    }
    
    // Don't modify if there was an error
    if ctx.Result().Error != nil {
        return nil
    }
    
    // Modify response from "Hello World" to "Modified response"
    ctx.Result().Data = "Modified response"
    return nil
})
```

### Database Hooks

#### OnPreDBQuery
```go
type PreDBQuery = func(
    ctx context.Context,
    option *QueryOption,
) error

func (a *App) OnPreDBQuery(hooks ...db.PreDBQuery)
```

Executed before database Get, First, Only operations. Can manipulate query before execution.

#### OnPostDBQuery
```go
type PostDBQuery = func(
    ctx context.Context,
    option *QueryOption,
    entities []*schema.Entity,
) ([]*schema.Entity, error)

func (a *App) OnPostDBQuery(hooks ...db.PostDBGet)
```

Executed after database Get, First, Only operations. Can manipulate query results:

```go
app.OnPostDBQuery(func(
    ctx context.Context,
    query *db.QueryOption,
    entities []*schema.Entity,
) ([]*schema.Entity, error) {
    if query.Model.Schema().Name != "file" {
        return entities, nil
    }
    
    // Add URL field to file entities
    for _, entity := range entities {
        path := entity.GetString("path")
        if path != "" {
            entity.Set("url", app.Disk().URL(path))
        }
    }
    return entities, nil
})
```

#### PreDBCreate / PostDBCreate
```go
type PreDBCreate = func(
    ctx context.Context,
    schema *schema.Schema,
    createData *schema.Entity,
) error

type PostDBCreate = func(
    ctx context.Context,
    schema *schema.Schema,
    dataCreate *schema.Entity,
    id uint64,
) error
```

Pre/post hooks for Create operations. PreDBCreate can manipulate data before creation, PostDBCreate can trigger actions after creation:

```go
app.OnPostDBCreate(func(
    ctx context.Context,
    schema *schema.Schema,
    dataCreate *schema.Entity,
    id uint64,
) error {
    if schema.Name != "file" {
        return nil
    }
    
    // Handle file entity creation
    // e.g., move file to different location
    // or optimize file for faster access
    return nil
})
```

#### PreDBUpdate / PostDBUpdate
```go
type PreDBUpdate = func(
    ctx context.Context,
    schema *schema.Schema,
    predicates []*Predicate,
    updateData *schema.Entity,
) error

type PostDBUpdate = func(
    ctx context.Context,
    schema *schema.Schema,
    predicates []*Predicate,
    updateData *schema.Entity,
    originalEntities []*schema.Entity,
    affected int,
) error
```

Pre/post hooks for Update operations:

```go
app.OnPostDBUpdate(func(
    ctx context.Context,
    schema *schema.Schema,
    predicates []*Predicate,
    updateData *schema.Entity,
    originalEntities []*schema.Entity,
    affected int,
) error {
    if schema.Name != "tag" {
        return nil
    }
    
    // Handle tag entity updates
    // e.g., update tag in search index
    // or update tag in cache
    return nil
})
```

#### PreDBDelete / PostDBDelete
```go
type PreDBDelete = func(
    ctx context.Context,
    schema *schema.Schema,
    predicates []*Predicate,
) error

type PostDBDelete = func(
    ctx context.Context,
    schema *schema.Schema,
    predicates []*Predicate,
    originalEntities []*schema.Entity,
    affected int,
) error
```

Pre/post hooks for Delete operations:

```go
app.OnPostDBDelete(func(
    ctx context.Context,
    schema *schema.Schema,
    predicates []*Predicate,
    originalEntities []*schema.Entity,
    affected int,
) error {
    if schema.Name != "tag" {
        return nil
    }
    
    // Handle tag entity deletion
    // e.g., remove tag from search index
    // or remove tag from cache
    return nil
})
```

#### PreDBExec / PostDBExec
```go
type PreDBExec = func(
    ctx context.Context,
    option *QueryOption,
) error

type PostDBExec = func(
    ctx context.Context,
    option *QueryOption,
    result sql.Result,
) error
```

Pre/post hooks for raw SQL Exec operations. Can manipulate queries or handle execution results.

## Advanced Storage System

FastSchema provides a powerful filesystem abstraction using **[rclone](https://rclone.org/)** integration, offering straightforward drivers for various storage systems.

### Storage Architecture

The **[rclonefs](https://pkg.go.dev/github.com/fastschema/fastschema/pkg/rclonefs)** package provides filesystem abstraction for reading and writing files across different storage systems.

### Supported Storage Systems
- **Local filesystem**: Default local storage
- **S3 compatible storage**: AWS S3, MinIO, DigitalOcean Spaces
- **Future support**: Many more storage systems planned

### Disk Interface

A Disk represents a filesystem for specific storage systems:

```go
type Disk interface {
    Name() string
    Root() string
    URL(filepath string) string
    Delete(c context.Context, filepath string) error
    Put(c context.Context, file *File) (*File, error)
    PutReader(c context.Context, in io.Reader, size uint64, mime, dst string) (*File, error)
    PutMultipart(c context.Context, m *multipart.FileHeader, dsts ...string) (*File, error)
    LocalPublicPath() string
}
```

### Storage Configuration

#### Configuration Structures
```go
type StorageConfig struct {
    DefaultDisk string        `json:"default_disk"`
    Disks       []*DiskConfig `json:"disks"`
}

type DiskConfig struct {
    Name            string `json:"name"`
    Driver          string `json:"driver"`
    Root            string `json:"root"`
    BaseURL         string `json:"base_url"`
    PublicPath      string `json:"public_path"`
    Provider        string `json:"provider"`
    Endpoint        string `json:"endpoint"`
    Region          string `json:"region"`
    Bucket          string `json:"bucket"`
    AccessKeyID     string `json:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key"`
    ACL             string `json:"acl"`
}
```

#### Default Disk Configuration
```json
{
  "name": "public",
  "driver": "local",
  "root": "./public",
  "public_path": "/files",
  "base_url": "http://localhost:8000/files"
}
```

#### Configuration Methods

1. **Environment Variables**:
   ```bash
   STORAGE='{"default_disk":"public","disks":[{"name":"public","driver":"local","root":"./public"}]}'
   ```

2. **Application Configuration**:
   ```go
   app.Config.StorageConfig = &fs.StorageConfig{
       DefaultDisk: "local_public",
       Disks: []*fs.DiskConfig{
           {
               Name:       "local_public",
               Driver:     "local",
               Root:       "./public",
               BaseURL:    "http://localhost:8000/files",
               PublicPath: "/files",
           },
       },
   }
   ```

### Storage Usage

#### Creating Storage Disks
```go
disks, err := rclonefs.NewFromConfig(
    storageDisksConfig, // []*DiskConfig
    localRoot,          // string representing local root path
)
```

#### Using Storage Disks
```go
// Get the default disk
disk := app.Disk()

// Get a specific disk
awsS3Disk := app.Disk("awss3")

// Get disk information
diskName := disk.Name()
diskRoot := disk.Root()
diskURL := disk.URL("path/to/file")

// Put a file
file, err := disk.Put(c, &fs.File{
    Name:   "file.txt",
    Path:   "path/to/file.txt",
    Type:   "text/plain",
    Size:   1024,
    Reader: strings.NewReader("Hello, World!"),
})

// Put from reader
file, err := disk.PutReader(
    ctx,
    strings.NewReader("content"),
    7,
    "text/plain",
    "path/to/file.txt",
)

// Put multipart file
file, err := disk.PutMultipart(ctx, fileHeader, "custom/path.txt")

// Delete file
err := disk.Delete(ctx, "path/to/file.txt")
```

#### Advanced Storage Examples

**Multiple Disk Configuration**:
```go
app := utils.Must(fastschema.New(&fs.Config{
    StorageConfig: &fs.StorageConfig{
        DefaultDisk: "local_public",
        Disks: []*fs.DiskConfig{
            {
                Name:       "local_public",
                Driver:     "local",
                Root:       "./public",
                BaseURL:    "http://localhost:8000/files",
                PublicPath: "/files",
            },
            {
                Name:            "s3_storage",
                Driver:          "s3",
                Provider:        "AWS",
                Region:          "us-west-2",
                Bucket:          "my-bucket",
                AccessKeyID:     "ACCESS_KEY",
                SecretAccessKey: "SECRET_KEY",
                BaseURL:         "https://my-bucket.s3.amazonaws.com",
            },
        },
    },
}))
```

**File Operations with Different Disks**:
```go
// Use default disk
publicDisk := app.Disk()
publicFile := utils.Must(publicDisk.Put(ctx, &fs.File{
    Name:   "public-file.txt",
    Path:   "uploads/public-file.txt",
    Type:   "text/plain",
    Size:   11,
    Reader: bytes.NewReader([]byte("Hello world")),
}))

// Use S3 disk
s3Disk := app.Disk("s3_storage")
s3File := utils.Must(s3Disk.Put(ctx, &fs.File{
    Name:   "s3-file.txt",
    Path:   "backups/s3-file.txt",
    Type:   "text/plain",
    Size:   13,
    Reader: bytes.NewReader([]byte("S3 Hello world")),
}))
```

This documentation provides a comprehensive overview of FastSchema for LLM understanding, covering architecture, usage patterns, configuration, development workflows, plugin systems, dashboard architecture, comprehensive hooks system, advanced storage management, and practical coding examples.
