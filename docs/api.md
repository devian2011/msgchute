# 📨 Notification Service API

**Version:** 1.0 **Base path:** `/`

This service provides an HTTP interface for sending notifications via various transports (email, SMS, messengers, etc.) with support for templates, retries, and administrative management.

**🔐 Authentication**  
All administrative endpoints (prefix `/api/admin/`) require a **Bearer token** in the `Authorization` header:

Authorization: Bearer <token>

- - -

## 📬 Public Endpoints

POST `/api/v1/send`

**Send a single message** — accepts a full message payload, queues it for delivery, and returns the created message and its associated task.

#### Request Body (application/json)

Object `SenderMessageRequest` (or `Message`) – see structure below.

| Field | Type | Description |
| --- | --- | --- |
| `sender_id` | string | Sender identifier (required) |
| `recipients` | \[\]string | List of recipients (email, phone, etc.) (required) |
| `subject` | string | Message subject |
| `body` | string | Message body (may contain template placeholders) |
| `code` | string | Template code if using a predefined template |
| `params` | object | Parameters for template substitution (key → value) |
| `transport` | string | Delivery transport (e.g., `email`, `sms`) (required) |
| `schedule` | string (datetime) | Scheduled send time (RFC3339) |
| `deadline` | string (datetime) | Deadline for processing |
| `meta` | object | Additional metadata (CC, BCC, attachments, etc.) |
| `retry` | object | Retry settings (`retries`, `strategy`, `params`) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Successfully processed | `AddMessageResponse` (fields: `message`, `task`) |
| 400 | Invalid JSON or structure | `Response` |
| 500 | Internal server error | `Response` |

POST `/api/v1/batch/send`

**Send multiple messages** — accepts an array of message payloads, processes each, and returns a list of results (one per message). Each entry contains the created message and its task, plus an optional error field.

#### Request Body (application/json)

Array of `SenderMessageRequest` objects (same structure as single send).

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Successfully processed | `[]AddBatchMessageResponse` (array of objects, each with `message`, `task`, and optional `error`) |
| 400 | Invalid JSON or structure | `Response` |
| 500 | Internal server error | `Response` |

POST `/api/v1/message/retry`

**Retry a message** — creates a new task to retry a previously failed or pending message, optionally overriding retry policy and schedule.

#### Request Body (application/json)

Object `MessageRetryRequest`:

| Field | Type | Description |
| --- | --- | --- |
| `id` | string (UUID) | ID of the message to retry (required) |
| `schedule` | string (datetime) | Desired time to run the new task |
| `deadline` | string (datetime) | New deadline for processing |
| `retry` | object | Override retry policy (`retries`, `strategy`, `params`) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Retry task created | `AddMessageResponse` (contains original message and new task) |
| 400 | Invalid request payload | `Response` |
| 500 | Internal server error | `Response` |

POST `/api/v1/preview`

**Preview a message** — renders a template with the given parameters and returns the final subject and body (without sending).

#### Request Body (application/json)

Object `PreviewMessageRequest`:

| Field | Type | Description |
| --- | --- | --- |
| `code` | string | Template code (if using a stored template) |
| `subject` | string | Subject override |
| `body` | string | Body override |
| `params` | object | Parameters for substitution (key → value) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Template rendered successfully | `MessagePreview` (fields: `subject`, `body`) |
| 400 | Invalid JSON | `Response` |
| 500 | Internal server error | `Response` |

- - -

## 🛠️ Administrative Endpoints

_All admin endpoints require Bearer token authentication._

GET `/api/admin/v1/message`

**List messages** — returns a paginated list of messages with filtering by status, IDs, recipients, senders, codes, or transports.

#### Query Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `page` | integer | Page number (default 1) |
| `per_page` | integer | Items per page (default 10) |
| `sort` | string | Sort field |
| `order` | string | Sort order (`asc` or `desc`) |
| `status` | \[\]string | Filter by status (repeatable) |
| `ids` | \[\]string (UUID) | Filter by message UUIDs |
| `recipient` | \[\]string | Filter by recipients |
| `sender_ids` | \[\]string | Filter by sender IDs |
| `code` | \[\]string | Filter by template codes |
| `transport` | \[\]string | Filter by transports |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | List of messages | `MessageFinderResponse` (fields: `messages` – array of `FullMessageInfo`, and `pagination`) |
| 400 | Invalid query parameters | `Response` |
| 500 | Internal server error | `Response` |

GET `/api/admin/v1/message/{id}`

**Get message by ID** — returns full details for a specific message, including all associated tasks and their results.

#### Path Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `id` | string (UUID) | Unique message identifier (required) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Detailed message information | `FullMessageInfo` (contains `message` and array of `tasks`) |
| 400 | Invalid UUID format | `Response` |
| 500 | Internal server error | `Response` |

GET `/api/admin/v1/template`

**List templates** — returns a paginated list of templates with filtering by code or full‑text search.

#### Query Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `page` | integer | Page number (default 1) |
| `per_page` | integer | Items per page (default 10) |
| `sort` | string | Sort field |
| `order` | string | Sort order (`asc` or `desc`) |
| `code` | \[\]string | Filter by template codes (repeatable) |
| `search` | string | Search phrase within template content |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | List of templates | `TemplateFinderResult` (contains `templates` – map of `{code: Template}`, and `pagination`) |
| 400 | Invalid query parameters | `Response` |
| 500 | Internal server error | `Response` |

POST `/api/admin/v1/template`

**Create a template** — stores a new message template in the system.

#### Request Body (application/json)

Object `Template`:

| Field | Type | Description |
| --- | --- | --- |
| `code` | string | Unique template code (required) |
| `name` | string | Template name (required) |
| `description` | string | Description |
| `subject` | string | Message subject (may contain placeholders) (required) |
| `body` | string | Message body (may contain placeholders) (required) |
| `params` | object | Expected parameters description (key → `TemplateParam` with `default` field) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Template created | `Template` (saved object) |
| 400 | Invalid JSON or structure | `Response` |
| 500 | Internal server error | `Response` |

PUT `/api/admin/v1/template/{code}`

**Update a template** — replaces an existing template identified by its code.

#### Path Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `code` | string | Template code to update (required) |

#### Request Body (application/json)

Object `Template` (same fields as creation).

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Template updated | `Template` (updated object) |
| 400 | Invalid JSON or structure | `Response` |
| 500 | Internal server error | `Response` |

- - -

## 📦 Core Data Structures

### Message

Main message object. Includes ID, sender, recipients, subject, body, template code, parameters, transport, schedule, deadline, metadata, retry settings, and status.

| Field | Type | Description |
| --- | --- | --- |
| `id` | string (UUID) | Unique message identifier |
| `sender_id` | string | Sender identifier |
| `recipients` | \[\]string | List of recipients |
| `subject` | string | Message subject |
| `body` | string | Message body |
| `code` | string | Template code |
| `params` | object | Template parameters (key → value) |
| `transport` | string | Transport type |
| `schedule` | string (datetime) | Scheduled send time |
| `deadline` | string (datetime) | Processing deadline |
| `meta` | object | Additional metadata |
| `retry` | object | Retry policy |
| `status` | string | Message status (enum) |

### Task

Delivery task linked to a message. Contains ID, message ID, status, retry count, creation time, last and next run times, worker identifier, and backoff parameters.

| Field | Type | Description |
| --- | --- | --- |
| `id` | string (UUID) | Task identifier |
| `message_id` | string (UUID) | Associated message ID |
| `worker` | string | Worker name (transport) |
| `status` | string | Task status (pending/success/failure) |
| `retries` | integer | Current retry count |
| `max_retries` | integer | Maximum retries allowed |
| `backoff_code` | string | Backoff strategy code |
| `backoff_params` | object | Parameters for backoff |
| `created_at` | string (datetime) | Creation timestamp |
| `last_run` | string (datetime) | Last execution timestamp |
| `next_run` | string (datetime) | Next scheduled run |
| `deadline` | string (datetime) | Task deadline |
| `is_processed` | boolean | Lock flag |

### Template

Message template. Defines code, name, description, subject, body, and expected parameters with default values.

| Field | Type | Description |
| --- | --- | --- |
| `code` | string | Unique template code (required) |
| `name` | string | Template name (required) |
| `description` | string | Description |
| `subject` | string | Subject template (required) |
| `body` | string | Body template (required) |
| `params` | object | Parameter definitions (key → `TemplateParam`) |

### FullMessageInfo

Extended message information that includes the message itself and an array of `FullTask` (tasks with their execution results).

### Response (error format)

Standard error response: contains `status`, `error` (error message), and optionally `data`.

### Status Enums

*   **MessageStatus**: `running`, `succeeded`, `failed`, `declined`
*   **TaskStatus**: `pending`, `success`, `failure`

Documentation generated from the service Swagger specification.
