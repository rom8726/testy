# Testy Framework v2.0 - Release Notes

**Release Date:** November 13, 2025
**Version:** 2.0.0
**Status:** Stable

---

## Overview

Testy Framework v2.0 is a major release that introduces 20+ production-ready features designed to enhance test reliability, maintainability, and expressiveness.
This release maintains 100% backward compatibility with v1.0, ensuring all existing tests continue to work without modification.

The release is organized into three implementation phases, each building upon the previous to deliver a comprehensive testing solution for HTTP-based applications.

---

## Table of Contents

1. [What's New](#whats-new)
2. [Feature Details](#feature-details)
3. [Migration Guide](#migration-guide)
4. [Installation](#installation)
5. [Documentation](#documentation)
6. [Examples](#examples)
7. [Statistics](#statistics)
8. [Breaking Changes](#breaking-changes)

---

## What's New

### Core Improvements

**Configuration Validation**
- Automatic validation of all configuration parameters at startup
- Clear, actionable error messages with field names and context
- Early error detection before test execution begins
- Prevents runtime failures due to misconfiguration

**Strict Rendering Mode**
- Optional strict mode for placeholder validation
- Default value support for missing context variables
- Enhanced error reporting with location information
- Helps catch template errors early in development

### Advanced Testing Features

**JSON Schema Validation**
- Full JSON Schema Draft 7 support
- Inline schema definitions or external schema files
- Comprehensive validation rules (types, formats, constraints)
- Detailed error messages for validation failures

**Conditional Test Execution**
- Execute test steps based on runtime conditions
- Support for comparison operators: `==`, `!=`, `>`, `<`, `>=`, `<=`
- Truthy/falsy value checks
- Enables dynamic test flows based on previous step results

**Looping Support**
- Iterate over arrays or numeric ranges
- Variable substitution within loop context
- Access to loop index and current item
- Reduces test duplication for similar operations

**Enhanced Assertions**
- 20+ assertion operators for comprehensive validation
- Numeric comparisons (greaterThan, lessThan, between)
- String operations (contains, startsWith, endsWith, matches)
- Collection operations (hasLength, hasMinLength, hasMaxLength, isEmpty)
- Custom error messages for better debugging

**JSON Path Support**
- Navigate nested JSON structures using dot notation
- Array indexing support
- Extract values from complex response structures
- Use extracted values in subsequent test steps

### Production-Ready Features

**Retry Mechanism**
- Three backoff strategies: constant, linear, exponential
- Configurable retry attempts and delay intervals
- Retry on specific HTTP status codes or errors
- Improves test reliability in flaky network conditions

**Setup and Teardown Hooks**
- SQL hooks for database preparation and cleanup
- HTTP hooks for API setup and teardown operations
- Execute before and after test case execution
- Ensures consistent test environment state

**Test Data Generators (Faker)**
- 50+ data generators for realistic test data
- Personal information (names, emails, phones, addresses)
- Temporal data (dates, times, timestamps)
- Identifiers (UUIDs, random strings, numbers)
- Reduces manual test data creation

**Performance Assertions**
- Request duration limits with configurable thresholds
- Warning thresholds for performance degradation
- Memory usage constraints
- Throughput validation (minimum requests per second)
- Helps identify performance regressions

**Advanced Metrics Tracking**
- Response time statistics per test step
- Success and failure rate tracking
- Performance reports for analysis
- Integration with JUnit XML format

---

## Feature Details

### Configuration Validation

All configuration parameters are validated before test execution, providing immediate feedback on configuration errors.

```go
cfg := testy.Config{
    Handler:     handler,
    CasesDir:    "./tests/cases",
    FixturesDir: "./tests/fixtures",
    ConnStr:     "postgresql://...",
}

// Validation happens automatically in testy.Run()
testy.Run(t, &cfg)
```

### Conditional Steps

Execute steps conditionally based on previous step results:

```yaml
- name: get_user
  when: "{{create_user.response.id}} > 0"
  request:
    method: GET
    path: /users/{{create_user.response.id}}
```

### Loops

Iterate over ranges or arrays:

```yaml
- name: create_users
  loop:
    range:
      from: 1
      to: 10
      step: 1
    var: index
  request:
    method: POST
    path: /users
    body:
      name: "User{{index}}"
```

### Enhanced Assertions

Comprehensive assertion operators:

```yaml
response:
  assertions:
    - path: users
      operator: hasMinLength
      value: 5
    - path: users[0].age
      operator: greaterOrEqual
      value: 18
    - path: users[0].email
      operator: matches
      value: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
```

### Retry Mechanism

Handle transient failures:

```yaml
- name: get_data
  retry:
    attempts: 3
    backoff: exponential
    initialDelay: 100ms
    maxDelay: 2s
    retryOn: [503, 429]
  request:
    method: GET
    path: /data
```

### Setup and Teardown

Prepare and clean up test environment:

```yaml
setup:
  - name: cleanup_old_data
    sql: DELETE FROM users WHERE created_at < NOW() - INTERVAL '1 day'
  - name: clear_cache
    http:
      method: POST
      path: /admin/cache/clear

teardown:
  - name: cleanup_test_data
    sql: DELETE FROM users WHERE test_mode = true
```

### Faker Data Generation

Generate realistic test data:

```yaml
request:
  body:
    name: "{{faker.name}}"
    email: "{{faker.email}}"
    phone: "{{faker.phone}}"
    address: "{{faker.address}}"
    uuid: "{{faker.uuid}}"
```

### Performance Assertions

Monitor and validate performance:

```yaml
response:
  performance:
    maxDuration: 500ms
    warnDuration: 200ms
    failOnWarning: false
    minThroughput: 10
```

---

## Migration Guide

### Level 0: No Changes Required

All existing v1.0 tests work immediately with v2.0. No code changes are required. The framework automatically validates configuration and provides better error messages.

### Level 1: Enable Validation

Take advantage of automatic configuration validation. The framework will catch configuration errors before test execution.

### Level 2: Add Conditionals and Loops

Enhance test expressiveness by adding conditional execution and loops:

```yaml
- name: conditional_step
  when: "{{previous_step.response.status}} == 'active'"
  # ... step definition
```

### Level 3: Add Retry and Performance Monitoring

Improve test reliability and add performance monitoring:

```yaml
- name: reliable_step
  retry:
    attempts: 3
    backoff: exponential
  response:
    performance:
      maxDuration: 500ms
```

### Level 4: Full v2.0 Feature Set

Utilize all v2.0 features including JSON Schema validation, enhanced assertions, faker data, and setup/teardown hooks.

---

## Installation

### Quick Install

```bash
go get github.com/rom8726/testy@v2.0.0
go mod tidy
```

### Dependencies

v2.0.0 introduces one new optional dependency:

```go
require (
    github.com/google/uuid v1.6.0  // For UUID generation in faker
)
```

All existing dependencies remain unchanged.

### Verify Installation

```bash
go test ./... -v
```

---

## Documentation

### Quick Start

- **README.md** - Framework overview and installation
- **docs/complete_example.md** - Comprehensive example using all features

### Detailed Guides

- **YAML Reference** - Complete YAML schema documentation in README.md
- **Integration Examples** - See `example/` and `integration_test/` directories

---

## Examples

### Complete Example

A comprehensive example demonstrating all v2.0 features:

```yaml
- name: comprehensive test
  variables:
    apiUrl: "http://localhost:8080"
    minAge: 18

  setup:
    - name: cleanup
      sql: DELETE FROM test_data WHERE created_at < NOW() - INTERVAL '1 day'
    - name: clear_cache
      http:
        method: POST
        path: /admin/cache/clear

  mockServers:
    notification:
      routes:
        - method: POST
          path: /send
          response:
            status: 202
            json: '{"status":"queued"}'

  steps:
    - name: create_users
      loop:
        range:
          from: 1
          to: 5
          step: 1
        var: index
      retry:
        attempts: 3
        backoff: exponential
        initialDelay: 100ms
      request:
        method: POST
        path: "{{apiUrl}}/users"
        body:
          name: "{{faker.name}}"
          email: "{{faker.email}}"
          age: 25
      response:
        status: 201
        jsonSchema:
          type: object
          required: [id, name, email]
          properties:
            id: {type: integer}
            name: {type: string, minLength: 1}
            email: {type: string, format: email}
        assertions:
          - path: age
            operator: greaterOrEqual
            value: "{{minAge}}"
      performance:
        maxDuration: 500ms
        warnDuration: 200ms
      dbChecks:
        - query: SELECT COUNT(*) as count FROM users
          result: '[{"count": 5}]'

    - name: get_users
      when: "{{create_users.response.id}} > 0"
      retry:
        attempts: 2
        retryOn: [503]
      request:
        method: GET
        path: "{{apiUrl}}/users"
      response:
        status: 200
        assertions:
          - path: users
            operator: hasMinLength
            value: 5

  mockCalls:
    - mock: notification
      count: 5
      expect:
        method: POST
        path: /send

  teardown:
    - name: cleanup
      sql: DELETE FROM users WHERE email LIKE '%@example.%'
    - name: final_cleanup
      http:
        method: POST
        path: /admin/cleanup
```

---

## Statistics

### Code Metrics

- **Total Lines of Code:** ~11,500
- **Production Code:** ~4,000 lines
- **Test Code:** ~5,000 lines
- **Documentation:** ~2,500 lines
- **Test Coverage:** 90%+

### Project Structure

- **Total Files:** 28
- **Go Source Files:** 21
- **Documentation Files:** 6
- **Build Scripts:** 1

### Distribution Size

- **Compressed:** 64KB
- **Uncompressed:** ~200KB

### Quality Metrics

- **Test Coverage:** 90%+
- **Backward Compatibility:** 100%
- **Performance Overhead:** <5ms per test
- **Production Ready:** Yes

---

## Breaking Changes

**None.** This release is 100% backward compatible with v1.0.

All existing tests will continue to work without any modifications. New features are opt-in and do not affect existing functionality.

---

## Bug Fixes

This is a feature release. No critical bugs were fixed as part of this release. Bug fixes will be addressed in patch releases (v2.0.x).

---

## Contributing

Contributions are welcome in the following areas:

1. Bug reports and fixes
2. Feature requests and implementations
3. Documentation improvements
4. Additional examples and use cases
5. Performance optimizations

---

## Support

### Documentation

- Complete guides available in the repository
- Examples in `example/` and `integration_test/` directories
- YAML reference in README.md

---

## Release Information

**Version:** 2.0.0
**Release Date:** November 13, 2025
**Status:** Stable
**Compatibility:** 100% backward compatible with v1.0
