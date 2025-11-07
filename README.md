# testy

[![Go Reference](https://pkg.go.dev/badge/github.com/rom8726/testy.svg)](https://pkg.go.dev/github.com/rom8726/testy)
[![Go Report Card](https://goreportcard.com/badge/github.com/rom8726/testy)](https://goreportcard.com/report/github.com/rom8726/testy)
[![Coverage Status](https://coveralls.io/repos/github/rom8726/testy/badge.svg?branch=main)](https://coveralls.io/github/rom8726/testy?branch=main)

[![boosty-cozy](https://gideonwhite1029.github.io/badges/cozy-boosty_vector.svg)](https://boosty.to/dev-tools-hacker)

A tiny functional-testing framework for HTTP handlers written in Go.
It lets you describe end-to-end scenarios in YAML, automatically:

* runs the request against the given `http.Handler`
* asserts the HTTP response (status code + JSON body)
* loads database fixtures before the scenario
* executes SQL checks after each step
* spins up lightweight HTTP mocks and verifies outbound calls

Only PostgreSQL and MySQL are supported.

---

## Why another testing tool?

* Write tests in plain YAML — easy for both developers and QA.
* Works with *any* `http.Handler` — net/http, Gin, Chi, Echo, ...
* Context-aware templating (response + env) out of the box.
* Uses libraries for JSON assertions and data fixtures.

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
 ├─ api/                 # your application code
 ├─ tests/
 │   ├─ cases/           # *.yml files with scenarios
 │   ├─ fixtures/        # *.yml fixtures for pgfixtures
 └── api_test.go         # Go test that calls testy.Run
```
### 1. Multi-step YAML case (`tests/cases/user_flow.yml`)
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

### 2. Go test (`tests/api_test.go`)
```go
package project_test

import (
  "os"
  "testing"

  "project/api"
  "github.com/rom8726/testy"
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

Run the tests:
```bash
go test ./...
```

---

## Features

### Declarative scenarios
* Unlimited **steps** per scenario.
* Each step can:
    * send any HTTP method, URL, headers and JSON body
    * reference values from previous responses (`{{<step>.response.<field>}}`)
    * inject environment variables (`{{HOME}}`, `{{UUID}}`, ...)
    * assert status code and JSON body (via [`kinbiko/jsonassert`](https://github.com/kinbiko/jsonassert))
    * run one or more **DB checks** — SQL + expected JSON\YAML rows.
    * fire and assert HTTP **mocks** for outgoing calls

### PostgreSQL fixtures
Loaded with [`rom8726/pgfixtures`](https://github.com/rom8726/pgfixtures):

* One YML file per table (or group of tables)
* Auto-truncate and sequence reset before inserting

### Hooks
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
        path:    string            # placeholders {{...}} allowed
        headers:                   # optional
          X-Token: "{{TOKEN}}"
        body:                      # optional, any YAML\JSON
          userId: "123"

      response:
        status: integer
        headers:
          Content-Type: application/json
        json:   string             # optional, must be valid JSON

      dbChecks:                    # optional, list
        - query:  SQL string       # placeholders {{...}} allowed
          result: JSON|YAML        # expected rows as JSON array
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
- Types/requirements for `name`, `fixtures`, `mockServers`, `mockCalls`, `steps[*].request`, `steps[*].response`, `steps[*].dbChecks`.
- Enumerations for HTTP methods (`GET`, `POST`, ...); status code range `100..599`.
- Body fields allow placeholders similar to Testy runtime behavior.

---

## License

Apache-2.0 License

