# 📨 Notification Service API

**Version:** 1.0

**Base path:** `/`

This service provides an HTTP interface for sending notifications via various transports (email, SMS, messengers, etc.) with support for templates, retries, and administrative management.

## 🔐 Authentication

All administrative endpoints (prefix `/api/admin/`) require a **Bearer token** in the `Authorization` header.

```
Authorization: Bearer <token>
```

---

## 📬 Public Endpoints

### POST `/api/v1/send`

**Send a message** — accepts a full message payload, queues it for delivery, and returns the created message and its associated task.

#### Request Body (application/json)

Object `github_com_devian2011_msgchute_internal_dto.Message` with the following fields:

| Field | Type | Description |
| --- | --- | --- |
| `id` | string | Unique message identifier (auto‑generated if omitted) |
| `sender_id` | string | Sender identifier (required) |
| `recipients` | \[\]string | List of recipients (email, phone, etc.) |
| `subject` | string | Message subject |
| `body` | string | Message body (may contain template placeholders) |
| `code` | string | Template code if using a predefined template |
| `params` | object | Parameters for template substitution (key → value) |
| `transport` | string | Delivery transport (e.g., `email`, `sms`) |
| `schedule` | string | Scheduled send time (RFC3339 format) |
| `deadline` | string | Deadline for processing |
| `metadata` | object | Additional metadata (CC, BCC, attachments, etc.) |
| `retry` | object | Retry settings (`retries`, `strategy`, `params`) |
| `status` | string | Message status (set automatically) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Successfully processed | `internal_io_web_endpoint_public.SenderResponse` (contains `message` and `task`) |
| 400 | Invalid JSON or structure | `github_com_devian2011_msgchute_pkg_http_response.Response` |
| 500 | Internal server error | same |

### POST `/api/v1/preview`

**Preview a message** — renders a template with the given parameters and returns the final subject and body (without sending).

#### Request Body (application/json)

Object `internal_io_web_endpoint_public.PreviewMessageRequest`:

| Field | Type | Description |
| --- | --- | --- |
| `code` | string | Template code (if using a stored template) |
| `subject` | string | Subject override |
| `body` | string | Body override |
| `params` | object | Parameters for substitution (key → value) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Template rendered successfully | `github_com_devian2011_msgchute_internal_dto.MessagePreview` (fields `subject`, `body`) |
| 400 | Invalid JSON | `Response` |
| 500 | Internal server error | `Response` |

---

## 🛠️ Administrative Endpoints

### GET `/api/admin/v1/message`

**List messages** — returns a paginated list of messages with filtering by status, IDs, recipients, senders, codes, or transports.

#### Query Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `page` | integer | Page number (default 1) |
| `per_page` | integer | Items per page (default 10) |
| `sort` | string | Sort field |
| `order` | string | Sort order (`asc` or `desc`) |
| `status` | \[\]string | Filter by status (repeatable) |
| `ids` | \[\]string | Filter by message UUIDs |
| `recipient` | \[\]string | Filter by recipients |
| `sender_ids` | \[\]string | Filter by sender IDs |
| `code` | \[\]string | Filter by template codes |
| `transport` | \[\]string | Filter by transports |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | List of messages | `internal_io_web_endpoint_admin.MessageFinderResponse` (fields `messages` – array of `FullMessageInfo`, and `pagination`) |
| 400 | Invalid query parameters | `Response` |
| 500 | Internal server error | `Response` |

### GET `/api/admin/v1/message/{id}`

**Get message by ID** — returns full details for a specific message, including all associated tasks and their results.

#### Path Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `id` | string (UUID) | Unique message identifier |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Detailed message information | `github_com_devian2011_msgchute_internal_dto.FullMessageInfo` (contains `message` and array of `tasks`) |
| 400 | Invalid UUID format | `Response` |
| 500 | Internal server error | `Response` |

### GET `/api/admin/v1/template`

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
| 200 | List of templates | `internal_io_web_endpoint_admin.TemplateFinderResult` (contains `templates` – map of `{code: Template}`, and `pagination`) |
| 400 | Invalid query parameters | `Response` |
| 500 | Internal server error | `Response` |

### POST `/api/admin/v1/template`

**Create a template** — stores a new message template in the system.

#### Request Body (application/json)

Object `github_com_devian2011_msgchute_internal_dto.Template`:

| Field | Type | Description |
| --- | --- | --- |
| `code` | string | Unique template code (required) |
| `name` | string | Template name |
| `description` | string | Description |
| `subject` | string | Message subject (may contain placeholders) |
| `body` | string | Message body (may contain placeholders) |
| `params` | object | Expected parameters description (key → `TemplateParam` with `default` field) |

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Template created | `Template` (saved object) |
| 400 | Invalid JSON or structure | `Response` |
| 500 | Internal server error | `Response` |

### PUT `/api/admin/v1/template/{code}`

**Update a template** — replaces an existing template identified by its code.

#### Path Parameters

| Parameter | Type | Description |
| --- | --- | --- |
| `code` | string | Template code to update |

#### Request Body (application/json)

Object `Template` (same fields as creation).

#### Responses

| Code | Description | Schema |
| --- | --- | --- |
| 200 | Template updated | `Template` (updated object) |
| 400 | Invalid JSON or structure | `Response` |
| 500 | Internal server error | `Response` |

---

## 📦 Core Data Structures

#### `Message`

Main message object. Includes ID, sender, recipients, subject, body, template code, parameters, transport, schedule, deadline, metadata, retry settings, and status.

#### `Task`

Delivery task linked to a message. Contains ID, message ID, status, retry count, creation time, last and next run times, worker identifier, and backoff parameters.

#### `Template`

Message template. Defines code, name, description, subject, body, and expected parameters with default values.

#### `FullMessageInfo`

Extended message information that includes the message itself and an array of `FullTask` (tasks with their execution results).

#### `Response` (error format)

Standard error response: contains `status`, `error` (error message), and optionally `data`.

### Status Enums

-   **MessageStatus**: `running`, `succeeded`, `failed`, `declined`
-   **TaskStatus**: `pending`, `success`, `failure`

---

Documentation generated from the service Swagger specification.