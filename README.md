# testy

[![Go Reference](https://pkg.go.dev/badge/github.com/rom8726/testy.svg)](https://pkg.go.dev/github.com/rom8726/testy)
[![Go Report Card](https://goreportcard.com/badge/github.com/rom8726/testy)](https://goreportcard.com/report/github.com/rom8726/testy)
[![Coverage Status](https://coveralls.io/repos/github/rom8726/testy/badge.svg?branch=main)](https://coveralls.io/github/rom8726/testy?branch=main)

[![boosty-cozy](https://gideonwhite1029.github.io/badges/cozy-boosty_vector.svg)](https://boosty.to/dev-tools-hacker)

A functional-testing framework for HTTP handlers and gRPC services written in Go.
It lets you describe end-to-end scenarios in YAML, automatically:

* runs HTTP requests against the given `http.Handler` or gRPC calls against gRPC services
* asserts HTTP responses (status code + JSON body) or gRPC responses (status code + protobuf message)
* loads database fixtures before the scenario
* executes SQL checks after each step
* spins up lightweight HTTP mocks and verifies outbound calls
* supports conditional execution, loops, retries, and performance monitoring (v2.0)
* validates responses with JSON Schema (v2.0)
* generates realistic test data with faker functions (v2.0)

Only PostgreSQL and MySQL are supported.

<img src="docs/logo.png" width="500" alt="Testy logo">

---

## Table of Contents

- [Why another testing tool?](#why-another-testing-tool)
- [Installation](#installation)
- [Quick example](#quick-example)
  - [1. Multi-step YAML case (tests/cases/user_flow.yml)](#1-multi-step-yaml-case-testscasesuser_flowyml)
  - [2. Go test (tests/api_test.go)](#2-go-test-testsapitestgo)
- [Features](#features)
  - [Declarative scenarios](#declarative-scenarios)
  - [PostgreSQL fixtures](#postgresql-fixtures)
  - [Hooks](#hooks)
  - [HTTP mocks (quick glance)](#http-mocks-quick-glance)
  - [Zero reflection magic](#zero-reflection-magic)
- [YAML reference](#yaml-reference)
- [YAML schema for scenarios (IDE support)](#yaml-schema-for-scenarios-ide-support)
  - [GoLand / IntelliJ IDEA (JetBrains)](#goland--intellij-idea-jetbrains)
  - [VS Code](#vs-code)
  - [What the schema covers](#what-the-schema-covers)
- [GoLand plugin](#goland-plugin)
- [License](#license)

## Why another testing tool?

* Write tests in plain YAML — easy for both developers and QA.
* Works with *any* `http.Handler` — net/http, Gin, Chi, Echo, ...
* Context-aware templating (response + env) out of the box.
* Uses libraries for JSON assertions and data fixtures.

---

## Installation
```bash
go get github.com/rom8726/testy/v2@latest
```

Add the package under test (your web-application) to `go.mod` as usual.

---

## Quick example

Directory layout:

```
/project
 ├─ api/                 # your application code
 ├─ tests/
 │   ├─ cases/           # *.testy.yml files with scenarios
 │   ├─ fixtures/        # *.yml fixtures for pgfixtures
 └── api_test.go         # Go test that calls testy.Run
```
### 1. Multi-step YAML case for HTTP (`tests/cases/user_flow.testy.yml`)
```yaml
- name: end-to-end user flow
  fixtures:
    - users

  steps:
    - name: create_user
      request:
        method: POST
        path:   /users
        body:
          name:  "Alice"
          email: "alice@example.com"
      response:
        status: 201
        headers:
          Content-Type: application/json
        json: |
          {
            "id":        "<<PRESENCE>>",
            "name":      "Alice",
            "email":     "alice@example.com"
          }

    - name: get user
      request:
        method: GET
        path:   /users/{{create_user.response.id}}     # pulls "id" from the previous response
      response:
        status: 200
        headers:
          Content-Type: application/json
        json: |
          {
            "id":   "{{create_user.response.id}}",
            "name": "Alice"
          }

    - name: update email
      request:
        method: PATCH
        path:   /users/{{create_user.response.id}}
        headers:
          X-Request-Id: "{{UUID}}"            # env-var substitution
        body:
          email: "alice+new@example.com"
      response:
        status: 204
      dbChecks:
        - query:  SELECT email FROM users WHERE id = {{create_user.response.id}}
          result: '[{ "email":"alice+new@example.com" }]'
```

> *NOTE:* you should point the Content-Type header in the response section to right parsing.

How the placeholders work:

* `{{<step name>.response.<json path>}}` — value from a previous response.
  The JSON path uses dots for objects and `[index]` for arrays (`items[0].id`).
* `{{ENV_VAR}}` — replaced with the value of an environment variable available at test run-time.

Placeholders are resolved in the request URL, headers and body, as well as inside `dbChecks.query`.

### 2. Multi-step YAML case for gRPC (`tests/cases/user_grpc.testy.yml`)
```yaml
- name: gRPC user service test
  fixtures:
    - users

  variables:
    defaultEmail: "test@example.com"

  steps:
    - name: create_user
      grpcRequest:
        service: users.UserService
        method: CreateUser
        message:
          name: "John Doe"
          email: "{{defaultEmail}}"
          age: 30
        metadata:
          authorization: "Bearer token123"
          x-request-id: "{{UUID}}"
      grpcResponse:
        code: OK
        message: |
          {
            "id": "<<PRESENCE>>",
            "name": "John Doe",
            "email": "test@example.com",
            "age": 30
          }
      dbChecks:
        - query: SELECT COUNT(*) as count FROM users WHERE email = 'test@example.com'
          result: '[{"count": 1}]'

    - name: get_user
      grpcRequest:
        service: users.UserService
        method: GetUser
        message:
          id: "{{create_user.response.id}}"
      grpcResponse:
        code: OK
        message: |
          {
            "id": "{{create_user.response.id}}",
            "name": "John Doe",
            "email": "test@example.com",
            "age": 30
          }
        assertions:
          - path: name
            operator: eq
            value: "John Doe"
```

### 3. Go test for HTTP (`tests/api_test.go`)
```go
package project_test

import (
  "os"
  "testing"

  "project/api"
  "github.com/rom8726/testy/v2"
)

func TestAPI(t *testing.T) {
  connStr := os.Getenv("TEST_DB") // postgres://user:password@localhost:5432/db?sslmode=disable

  testy.Run(t, &testy.Config{
    Handler:     api.Router(),    // your http.Handler
    CasesDir:    "./cases",
    FixturesDir: "./fixtures",
    ConnStr:     connStr,
  })
}
```

### 4. Go test for gRPC (`tests/grpc_test.go`)
```go
package tests

import (
  "context"
  "net"
  "path/filepath"
  "runtime"
  "testing"
  "time"

  _ "github.com/lib/pq"
  "github.com/rom8726/pgfixtures"
  "github.com/testcontainers/testcontainers-go"
  "github.com/testcontainers/testcontainers-go/modules/postgres"
  "github.com/testcontainers/testcontainers-go/wait"
  "google.golang.org/grpc"

  "github.com/rom8726/testy/v2"
  "yourproject/server"
  "yourproject/userspb"
)

func TestGRPCServer(t *testing.T) {
  ctx := context.Background()

  // Start PostgreSQL container with testcontainers
  postgresContainer, err := postgres.RunContainer(ctx,
    testcontainers.WithImage("postgres:16"),
    postgres.WithDatabase("db"),
    postgres.WithUsername("user"),
    postgres.WithPassword("password"),
    testcontainers.WithWaitStrategy(
      wait.ForLog("database system is ready to accept connections").
        WithOccurrence(2).
        WithStartupTimeout(30*time.Second),
    ),
  )
  if err != nil {
    t.Fatalf("failed to start postgres container: %v", err)
  }
  defer func() {
    if err := postgresContainer.Terminate(ctx); err != nil {
      t.Fatalf("failed to terminate postgres container: %v", err)
    }
  }()

  // Get connection string
  connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
  if err != nil {
    t.Fatalf("failed to get connection string: %v", err)
  }
  // Convert postgres:// to postgresql:// for compatibility with lib/pq
  if len(connStr) >= 11 && connStr[:11] == "postgres://" {
    connStr = "postgresql://" + connStr[11:]
  }

  // Apply migration
  _, testFile, _, _ := runtime.Caller(0)
  testDir := filepath.Dir(testFile)
  migrationPath := filepath.Join(testDir, "grpc_migration.sql")
  if err := applyMigration(ctx, connStr, migrationPath); err != nil {
    t.Fatalf("failed to apply migration: %v", err)
  }

  // Create gRPC service servers
  userService, err := server.NewUserServiceServer(connStr)
  if err != nil {
    t.Fatalf("failed to create user service: %v", err)
  }
  defer userService.Close()

  // Start gRPC server
  grpcServer := grpc.NewServer()
  userspb.RegisterUserServiceServer(grpcServer, userService)

  // Start listening on a random port
  listener, err := net.Listen("tcp", "localhost:0")
  if err != nil {
    t.Fatalf("Failed to create listener: %v", err)
  }

  grpcAddr := listener.Addr().String()

  // Start server in goroutine
  go func() {
    if err := grpcServer.Serve(listener); err != nil {
      t.Logf("gRPC server error: %v", err)
    }
  }()
  defer grpcServer.GracefulStop()

  // Get test directory paths
  casesDir := filepath.Join(testDir, "cases")
  fixturesDir := filepath.Join(testDir, "fixtures")

  casesDir, err = filepath.Abs(casesDir)
  if err != nil {
    t.Fatalf("failed to get absolute path for cases: %v", err)
  }
  fixturesDir, err = filepath.Abs(fixturesDir)
  if err != nil {
    t.Fatalf("failed to get absolute path for fixtures: %v", err)
  }

  // Run testy tests
  cfg := testy.Config{
    GRPCAddr:    grpcAddr,
    DBType:      pgfixtures.PostgreSQL,
    CasesDir:    casesDir,
    FixturesDir: fixturesDir,
    ConnStr:     connStr,
  }

  testy.Run(t, &cfg)
}
```

Run the tests:
```bash
go test ./...
```

---

## Features

### Declarative scenarios
* Unlimited **steps** per scenario.
* Each step can:
    * send HTTP requests (any method, URL, headers and JSON body) or gRPC calls (service, method, message, metadata)
    * reference values from previous responses (`{{<step>.response.<field>}}`)
    * inject environment variables (`{{HOME}}`, `{{UUID}}`, ...)
    * assert HTTP responses (status code + JSON body) or gRPC responses (status code + protobuf message)
    * run one or more **DB checks** — SQL + expected JSON\YAML rows.
    * fire and assert HTTP **mocks** for outgoing calls

### v2.0 New Features

**Conditional Test Execution**
* Execute steps conditionally based on runtime values
* Support for comparison operators (`==`, `!=`, `>`, `<`, `>=`, `<=`)
* Truthy/falsy checks for dynamic test flows

**Looping Support**
* Iterate over arrays or numeric ranges
* Reduce test duplication for similar operations
* Access loop index and current item in templates

**Enhanced Assertions**
* 20+ assertion operators for comprehensive validation
* Numeric comparisons, string operations, collection checks
* Custom error messages for better debugging

**JSON Schema Validation**
* Full JSON Schema Draft 7 support
* Inline or external schema files
* Comprehensive validation rules

**Retry Mechanism**
* Three backoff strategies: constant, linear, exponential
* Retry on specific status codes or errors
* Improves test reliability in flaky conditions

**Setup and Teardown Hooks**
* SQL hooks for database preparation and cleanup
* HTTP hooks for API setup operations
* Execute before and after test case execution

**Test Data Generators (Faker)**
* 50+ data generators for realistic test data
* Names, emails, phones, addresses, dates, UUIDs
* Reduces manual test data creation

**Performance Assertions**
* Request duration limits with warnings
* Throughput validation
* Memory usage constraints

**JSON Path Support**
* Navigate nested JSON structures
* Extract values from complex responses
* Use in subsequent test steps

### PostgreSQL fixtures
Loaded with [`rom8726/pgfixtures`](https://github.com/rom8726/pgfixtures):

* One YML file per table (or group of tables)
* Auto-truncate and sequence reset before inserting

### Request Hooks
Optional pre/post request hooks to stub time, clean caches, etc.:
```go
testy.Run(t, &testy.Config{
    // ...
    BeforeReq: func() error { /* do something */ return nil },
    AfterReq:  func() error { /* do something */ return nil },
})
```

### HTTP mocks (quick glance)

Lightly describe external services directly in the scenario, then verify how many times (and with what payload) your code called them.

```yaml
mockServers:
  notification:
    routes:
      - method: POST
        path: /send
        response:
          status: 202
        headers:
          Content-Type: application/json
        json: '{"status":"queued"}'

mockCalls:
  - mock: notification
    count: 1
    expect:
      method: POST
      path: /send
      body:
        contains: "Joseph"
```

```go
func TestServer(t *testing.T) {
    mocks, err := testy.StartMockManager("notification")
    if err != nil {
        t.Fatalf("mock start: %v", err)
    }
    defer mocks.StopAll()

    err = os.Setenv("NOTIFICATION_BASE_URL", mocks.URL("notification"))
    if err != nil {
        t.Fatalf("set env: %v", err)
    }
    defer os.Unsetenv("NOTIFICATION_BASE_URL")

    // ...

    cfg := testy.Config{
        MockManager: mocks,
        // ...
    }
}
```

### Zero reflection magic
The framework only needs:

* an `http.Handler` (for HTTP tests) or gRPC server address (for gRPC tests)
* PostgreSQL connection string
* paths to your YAML files

For gRPC tests, the framework automatically discovers service methods from protobuf files registered in the global protobuf registry.
Make sure to import your generated protobuf packages in your test file to enable service discovery.

---

## YAML reference
```yaml
- name: string

  variables:                 # optional, test-level variables
    key: value

  fixtures:                  # optional, order matters
    - fixture-file           # without ".yml" extension

  setup:                     # optional, runs before steps
    - name: string           # optional hook name
      sql: SQL string        # SQL query to execute
    - http:                  # HTTP request hook
        method: POST
        path: /admin/reset
        headers:
          X-Key: value

  teardown:                  # optional, runs after steps (even on failure)
    - sql: SQL string
    - http: ...

  steps:
    - name: string

      when: string           # optional, conditional execution
      # examples: "{{status}} == active", "{{age}} >= 18"

      loop:                  # optional, iterate over items or range
        items: [...]         # list of items to iterate
        var: itemName        # variable name for current item
        # OR
        range:               # numeric range
          from: 1
          to: 10
          step: 1            # optional, default 1

      retry:                 # optional, retry on failure
        attempts: 3          # max number of attempts
        backoff: exponential # constant | linear | exponential
        initialDelay: 100ms  # first retry delay
        maxDelay: 10s        # max delay for exponential
        retryOn: [503, 429]  # retry only on these status codes
        retryOnError: true   # retry on any error

      # HTTP request (use either request/response OR grpcRequest/grpcResponse)
      request:
        method:  GET | POST | PUT | PATCH | DELETE | ...
        path:    string      # placeholders {{...}} allowed
        headers:             # optional
          X-Token: "{{TOKEN}}"
        body:                # optional, any YAML\JSON
          userId: "123"
          name: "{{faker.name}}"        # faker generators supported
          email: "{{faker.email}}"
          # Available faker functions:
          # - {{faker.uuid}}, {{faker.uuid.v4}} (UUID)
          # - {{faker.name}}, {{faker.firstName}}, {{faker.lastName}}, {{faker.fullName}} (names)
          # - {{faker.email}}, {{faker.username}}, {{faker.domain}}, {{faker.url}}, {{faker.ipv4}} (internet)
          # - {{faker.phone}}, {{faker.phoneNumber}} (phone)
          # - {{faker.city}}, {{faker.street}}, {{faker.country}}, {{faker.zipCode}} (address)
          # - {{faker.date}}, {{faker.time}}, {{faker.timestamp}}, {{faker.now}} (date/time)
          # - {{faker.number}}, {{faker.integer}}, {{faker.float}}, {{faker.digit}} (numbers)
          # - {{faker.word}}, {{faker.words}}, {{faker.sentence}}, {{faker.paragraph}} (text)
          # - {{faker.company}}, {{faker.companyName}} (company)
          # - {{faker.random.string}}, {{faker.random.int}}, {{faker.random.bool}} (random)

      # HTTP response
      response:
        status: integer
        headers:
          Content-Type: application/json
        json: string         # optional, must be valid JSON

        schema: path/to/schema.json     # optional, external JSON Schema file

        jsonSchema:          # optional, inline JSON Schema
          type: object
          required: [id, name]
          properties:
            id: {type: integer}
            name: {type: string}

        assertions:          # optional, enhanced assertions
          - path: users[0].age          # JSON path (supports dot notation and array indexing)
            operator: greaterThan       # Available operators:
              # - equals, eq, == (equality check)
              # - notEquals, ne, != (inequality check)
              # - greaterThan, gt, > (numeric comparison)
              # - lessThan, lt, < (numeric comparison)
              # - greaterOrEqual, gte, >= (numeric comparison)
              # - lessOrEqual, lte, <= (numeric comparison)
              # - between (value must be between two numbers: [min, max])
              # - contains (string/array contains value)
              # - notContains (string/array does not contain value)
              # - matches (regex pattern matching)
              # - startsWith (string starts with value)
              # - endsWith (string ends with value)
              # - in (value is in array)
              # - notIn (value is not in array)
              # - isEmpty (value is empty)
              # - isNotEmpty (value is not empty)
              # - hasLength (exact length check)
              # - hasMinLength (minimum length check)
              # - hasMaxLength (maximum length check)
            value: 18                   # expected value (for between: [min, max] array)
            message: string             # optional, custom error message

      # gRPC request (use either request/response OR grpcRequest/grpcResponse)
      grpcRequest:
        service: string      # gRPC service name (e.g., "users.UserService"), supports placeholders
        method: string      # gRPC method name (e.g., "CreateUser", "GetUser"), supports placeholders
        message:            # request message as key-value map
          name: "John Doe"
          email: "{{faker.email}}"
          age: 30
        metadata:           # optional, gRPC metadata (headers)
          authorization: "Bearer {{TOKEN}}"
          x-request-id: "{{UUID}}"

      # gRPC response
      grpcResponse:
        code: OK | NOT_FOUND | INVALID_ARGUMENT | ...  # gRPC status code
        message: string    # optional, expected response as JSON (supports jsonassert placeholders like <<PRESENCE>>)
        metadata:          # optional, expected gRPC metadata
          x-trace-id: "value"
        assertions:        # optional, enhanced assertions (same as HTTP response assertions)
          - path: name
            operator: eq
            value: "John Doe"

      performance:           # optional, performance constraints
        maxDuration: 500ms   # max allowed duration
        warnDuration: 200ms  # warning threshold
        failOnWarning: false # fail test on warning
        maxMemory: 256       # max memory in MB
        minThroughput: 10    # minimum requests per second (for batch operations)

      dbChecks:              # optional, list
        - query: SQL string  # placeholders {{...}} allowed
          result: JSON|YAML  # expected rows as JSON array
```

---

## YAML schema for scenarios (IDE support)

The repository provides `testy.json` — a JSON Schema for Testy YAML scenarios. You can attach it in your IDE to get:

- key auto-completion
- live validation and error highlighting (types, required fields, HTTP method enums, etc.)

To enable automatic validation, create your scenario files with one of these extensions:

- `.testy.yml`
- `.testy.yaml`

These patterns are used in the examples below when mapping the schema.

### GoLand / IntelliJ IDEA (JetBrains)

1) Open: Preferences | Languages & Frameworks | Schemas and DTDs | JSON Schema.
2) Click "+" and choose "User schema".
3) Set the schema file to `testy.json` at the project root, or use the raw GitHub URL:
   `https://raw.githubusercontent.com/rom8726/testy/main/testy.json`.
4) In "Schema mappings" add file patterns, for example:
   - `*.testy.yml`
   - `*.testy.yaml`
   - (optional) the whole `tests/cases` folder
5) Apply settings. Validation and completion will work in your YAML scenario files.

Notes:
- JetBrains IDEs can apply JSON Schema to YAML files (not just JSON).
- The schema targets draft-07 (supported by IDEs by default).

### VS Code

1) Install the "YAML" extension (Red Hat) — it supports mapping JSON Schema to YAML.
2) Add a mapping in `.vscode/settings.json` (or in global Settings → search: yaml.schemas):

```json
{
  "yaml.schemas": {
    "./testy.json": [
      "**/*.testy.yml",
      "**/*.testy.yaml"
    ]
  }
}
```

Alternatively, use the URL:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/rom8726/testy/main/testy.json": [
      "**/*.testy.yml",
      "**/*.testy.yaml"
    ]
  }
}
```

After this, VS Code will validate and auto-complete fields in Testy YAML scenarios.

### What the schema covers

- File root is an array of scenarios.
- Types/requirements for `name`, `fixtures`, `mockServers`, `mockCalls`, `steps[*].request`, `steps[*].response`, `steps[*].grpcRequest`, `steps[*].grpcResponse`, `steps[*].dbChecks`.
- Enumerations for HTTP methods (`GET`, `POST`, ...); status code range `100..599`.
- Enumerations for gRPC status codes (`OK`, `NOT_FOUND`, `INVALID_ARGUMENT`, ...).
- Body fields allow placeholders similar to Testy runtime behavior.
- Support for both HTTP and gRPC test steps (mutually exclusive per step).

## GoLand plugin

Enhance your workflow in JetBrains IDEs with the Testy Tests Viewer plugin.

- Install from JetBrains Marketplace: Testy Tests Viewer
  https://plugins.jetbrains.com/plugin/28969-testy-tests-viewer
- Or install manually via “Install Plugin from Disk…” and select the testy-goland-plugin.zip

The plugin provides a dedicated tests viewer for Testy scenarios.

---

## License

Apache-2.0 License

