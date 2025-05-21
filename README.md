# testy

A tiny functional-testing framework for HTTP handlers written in Go.  
It lets you describe end-to-end scenarios in YAML, automatically:

* runs the request against the given `http.Handler`
* asserts the HTTP response (status code + JSON body)
* loads database fixtures before the scenario
* performs SQL checks after each step

Only PostgreSQL is supported.

---

## Installation
```bash
go get github.com/rom8726/testy@latest
```

Add the package under test (your web-application) to `go.mod` as usual.

---

## Quick example

Directory layout:
```
/project
 ├─ api/           # your application code
 ├─ tests/
 │   ├─ cases/     # *.yml files with scenarios
 │   └─ fixtures/  # *.yml fixtures for pgfixtures
 └─── api_test.go  # Go test that calls testy.Run
```
### 1. YAML test case (`tests/cases/user_get.yml`)
```yaml
- name: GET /users/:id
  fixtures:
   - users          # loads tests/fixtures/users.yml
  steps:
    - name: happy path
      request:
        method: GET
        path:  /users/42
      response:
        status: 200
        json: |
          {
            "id":        42,
            "name":      "John Doe",
            "email":     "<<PRESENCE>>"
          }
      dbChecks:
        - query:  SELECT COUNT(*) AS cnt FROM users WHERE id = 42
          result: '[{ "cnt": 1 }]'
```

### 2. Fixture (`tests/fixtures/users.yml`)
```yaml
users:
  - id: 42
    name: John Doe
    email: john@example.com
```

### 3. Go test (`tests/api_test.go`)
```go
package project_test

import (
  "net/http"
  "os"
  "testing"

  "project/api"
  "github.com/rom8726/testy"
)

func TestAPI(t *testing.T) {
  connStr := os.Getenv("TEST_DB") // e.g. "postgres://user:pass@localhost:5432/app_test?sslmode=disable"

  testy.Run(t, &testy.Config{
    Handler:     api.Router(),    // your http.Handler
    CasesDir:    "./cases",
    FixturesDir: "./fixtures",
    ConnStr:     connStr,
  })
}
```

Run the tests:
```bash
go test ./...
```

---

## Features

### Declarative scenarios

* Compose any number of **steps** inside a scenario.
* Each step may:
    * send any HTTP method, path, headers and JSON body
    * assert status code
    * assert JSON body with [`kinbiko/jsonassert`](https://github.com/kinbiko/jsonassert)  
      Use `<<PRESENCE>>`, `<<UNORDERED>>` etc.
    * run one or more **DB checks** — SQL query plus expected JSON result.

### PostgreSQL fixtures

Fixtures are loaded with [`rom8726/pgfixtures`](https://github.com/rom8726/pgfixtures):

* YML file per table (or group of tables)
* Auto-truncate + sequence reset before inserting

### Hooks

Optionally add pre/post request hooks to mutate state, mock time, etc.:
```go
testy.Run(t, &testy.Config{
    // ...
    BeforeReq: func() error { /* do something */ return nil },
    AfterReq:  func() error { /* do something */ return nil },
})
```

### Zero reflection magic

The framework only needs:

* an `http.Handler`
* PostgreSQL connection string
* paths to your YAML files

---

## YAML reference
```yaml
- name: string

  fixtures:            # optional, order matters
    - fixture-file     # without ".yml" extension

  steps:
    - name: string

      request:
        method:  GET | POST | PUT | PATCH | DELETE | ...
        path:    string
        headers:               # optional
          Authorization: Bearer xyz
        body:                  # optional, any YAML -> JSON

      response:
        status: integer
        json:   string         # optional, must be valid JSON

      dbChecks:                # optional, list
        - query:  SQL string
          result: JSON|YAML    # expected rows as JSON array
```

---

## Why another testing tool?

* Write tests in plain YAML — easy for QA/devs to co-author.
* Works with *any* `http.Handler` — standard library, Gin, Chi, Echo, etc.
