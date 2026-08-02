# Notifications service

## Template engine

[Pongo - Django like](https://github.com/flosch/pongo2)

## Configuration

### .env

```dotenv
APP_HTTP_ADDR=:8080 # Http port
APP_STORAGE_DSN=file://PWD/templates # Path for save templates into files
#APP_STORAGE_DSN=sqlite://PWD/templates.db # Path for save templates into sqlite
APP_TRANSPORTS_CFG_PATH=./config/transports.yaml # Transports configurations
```

### transports.yaml

Application supports many transports. You can use multiple smtp, file and other transports.

```yaml
smtp: # Block for smtp provider
  gmail: # Transport code
    host: smtp.gmail.com # Smtp host
    port: 587 # Smtp port
    user: user # Smtp user
    password: password # Smtp password
    from: Mr Mail<mail@example.com> # Smtp from field
file: # Block for file provider
  console: # Transport code
    path: /dev/stdout #Output file
  file: # Transport code
    path: ./test.txt # Output file
telegram:
  tg:
    token: # Bot token
```

## API

### Send without template

POST **/api/v1/send**

Request:
```json
{
  "messages": [
    {
      "transport": "console",
      "to": ["recepient@example.com"],
      "message": {
        "subject": "Some subject",
        "body": "<body><p>Some body</p></body>"
      }
    }
  ]
}
```
Response:
```json
{
  "data":
    [
      {
        "success":true,
        "error":null
      }
    ],
  "err":null
}
```

### Send with template

POST **/api/v1/send**

Request:
```json
{
  "messages": [
    {
      "transport": "file",
      "to": ["recepient@example.com"],
      "params": {"second": "Second param"},
      "code": "template code"
    }
  ]
}
```

Response:
```json
{
  "data":
    [
      {
        "success":true,
        "error":null
      }
    ],
  "err":null
}
```

### Transport list

GET **/api/v1/transports**

Response:
```json
{
  "data": ["ya","console","file"],
  "err":null
}
```

### Create template

POST **/api/v1/create**

Request:
```json
{
  "code": "template_code",
  "description": "Some template description",
  "params": {
    "first": {
      "default": "Default value for first param"
    }
  },
  "subject": "{{.first}} Subject",
  "body": "Body {{.first}} World"
}
```

Response:
```json
{
  "data": {
    "code": "test",
    "description": "some test",
    "params": {
      "first": {
        "options": null,
        "default": "Hello"
      }
    },
    "subject": "{{.first}} Subject",
    "body": "Body {{.first}} World"
  },
  "err": null
}
```

### Update template

POST **/api/v1/update**

Request:
```json
{
  "code": "template_code",
  "description": "Some template description",
  "params": {
    "first": {
      "default": "Default value for first param"
    }
  },
  "subject": "{{.first}} Subject",
  "body": "Body {{.first}} World"
}
```

Response:
```json
{
  "data": {
    "code": "test",
    "description": "some test",
    "params": {
      "first": {
        "options": null,
        "default": "Hello"
      }
    },
    "subject": "{{.first}} Subject",
    "body": "Body {{.first}} World"
  },
  "err": null
}
```

### Remove template

POST **/api/v1/remove**

Request:
```json
{
  "code": "template_code"
}
```

Response:
```json
{
  "data": null,
  "err": null
}
```

### Get template

POST **/api/v1/get**

Request:
```json
{
  "code": "template_code"
}
```

Response:
```json
{
  "data": {
    "code": "test",
    "description": "some test",
    "params": {
      "first": {
        "options": null,
        "default": "Hello"
      }
    },
    "subject": "{{.first}} Subject",
    "body": "Body {{.first}} World"
  },
  "err": null
}
```

### List template

GET **/api/v1/list**

Response:
```json
{
  "data": [
    {
      "code": "test",
      "description": "some test",
      "params": {
        "first": {
          "options": null,
          "default": "Hello"
        }
      },
      "subject": "{{.first}} Subject",
      "body": "Body {{.first}} World"
    }
  ],
  "err": null
}
```
