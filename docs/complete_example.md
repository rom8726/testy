## Complete Example

All features combined:

```yaml
- name: comprehensive e-commerce flow
  variables:
    adminToken: {{ADMIN_TOKEN}}
    minAge: 18

  # Setup
  setup:
    - sql: DELETE FROM orders WHERE test_mode = true
    - http:
        method: POST
        path: /admin/cache/clear

  steps:
    # Create users with faker and loop
    - name: create_users
      loop:
        range: {from: 1, to: 5}
        var: index
      retry:
        attempts: 3
        backoff: exponential
      request:
        method: POST
        path: {{API_URL}}/users
        body:
          name: "{{faker.name}}"
          email: "{{faker.email}}"
          phone: "{{faker.phone}}"
          age: 25
      response:
        status: 201
        jsonSchema:
          type: object
          required: [id, name, email]
          properties:
            id: {type: integer}
            name: {type: string}
            email: {type: string}
        assertions:
          - path: age
            operator: greaterOrEqual
            value: "{{minAge}}"
      performance:
        maxDuration: 300ms

    # Get all users
    - name: get_users
      retry:
        attempts: 2
        retryOn: [503]
      request:
        method: GET
        path: {{API_URL}}/users
      response:
        status: 200
        assertions:
          - path: users
            operator: hasMinLength
            value: 5
      performance:
        maxDuration: 500ms
      dbChecks:
        - query: SELECT COUNT(*) as count FROM users
          result: '[{"count": 5}]'

    # Conditional upgrade
    - name: upgrade_first_user
      when: "{{get_users.response.users[0].age}} >= 21"
      request:
        method: POST
        path: /api/users/{{get_users.response.users[0].id}}/premium
      response:
        status: 200

  # Teardown
  teardown:
    - sql: DELETE FROM users WHERE email LIKE '%@example.%'
    - http:
        method: DELETE
        path: /cache/invalidate
```
