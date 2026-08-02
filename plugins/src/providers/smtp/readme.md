# 📧 SMTP Provider Plugin

A plugin for sending emails via SMTP, designed to be loaded and managed by the **msgchute** plugin system.

## Overview

This plugin implements the `provider.Provider` interface and enables:

*   Sending plain‑text emails via any SMTP server
*   Support for CC and BCC recipients
*   File attachments (base64‑encoded)
*   Customisable sender address
*   Full MIME compliance (multipart/mixed for attachments)

## Building

Compile the plugin into a binary that can be discovered by the plugin manager:

```
go build -o smtp-provider .
```

## Configuration

The plugin expects a JSON configuration object with the following fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `host` | string | ✅   | SMTP server hostname (e.g. `smtp.gmail.com`) |
| `port` | int | ✅   | SMTP server port (e.g. `587` for TLS, `465` for SSL) |
| `username` | string | ✅   | SMTP authentication username |
| `password` | string | ✅   | SMTP authentication password |
| `from` | string | ❌   | Default sender email address (can be overridden per message) |

Example:

```
{
    "host": "smtp.example.com",
    "port": 587,
    "username": "user@example.com",
    "password": "securepassword",
    "from": "noreply@example.com"
}
```

## Usage

The plugin exposes the `Send(msg *provider.Message) error` method. The `Message` structure is defined as:

```
type Message struct {
    To      []string // list of recipient email addresses
    Subject string   // email subject
    Body    string   // plain‑text email body
    Params  []byte   // JSON-encoded SendOptions (see below)
}
```

### Send Options (`Params`)

The `Params` field must contain a JSON object with these optional properties:

| Field | Type | Description |
| --- | --- | --- |
| `from` | string | Override the sender email address |
| `cc` | array of strings | List of CC recipients |
| `bcc` | array of strings | List of BCC recipients |
| `reply_to` | string | Reply‑To header value |
| `attachments` | array | List of files to attach (see below) |

#### Attachment Object

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `filename` | string | ✅   | Name of the file |
| `content` | string | ✅   | Base64‑encoded file content |
| `content_type` | string | ❌   | MIME type (defaults to `application/octet-stream`) |

## Examples

### 1\. Simple email

```
msg := &provider.Message{
    To:      []string{"recipient@example.com"},
    Subject: "Hello",
    Body:    "This is a test email.",
    Params:  []byte(`{}`),
}
```

### 2\. Email with CC, BCC and custom sender

```
params := map[string]interface{}{
    "from":   "custom@example.com",
    "cc":     []string{"cc@example.com"},
    "bcc":    []string{"bcc@example.com"},
    "reply_to": "reply@example.com",
}
paramsJSON, _ := json.Marshal(params)

msg := &provider.Message{
    To:      []string{"to@example.com"},
    Subject: "Meeting",
    Body:    "Please confirm your attendance.",
    Params:  paramsJSON,
}
```

### 3\. Email with an attachment

```
import "encoding/base64"

fileContent := base64.StdEncoding.EncodeToString([]byte("Hello, world!"))
params := map[string]interface{}{
    "attachments": []map[string]string{
        {
            "filename":    "hello.txt",
            "content":     fileContent,
            "content_type": "text/plain",
        },
    },
}
paramsJSON, _ := json.Marshal(params)

msg := &provider.Message{
    To:      []string{"recipient@example.com"},
    Subject: "File attached",
    Body:    "Here is your file.",
    Params:  paramsJSON,
}
```

## Notes

*   The plugin uses `PLAIN` authentication over TLS. Make sure your SMTP server supports it.
*   Common ports: `587` (STARTTLS) and `465` (SSL/TLS). Adjust accordingly.
*   If the `from` field is not provided in config or overridden in `Params`, the `username` is used as the sender.
*   Attachments are sent as `base64` encoded parts. The plugin decodes them and builds a proper MIME multipart message.
*   All recipients (To, CC, BCC) are combined into the SMTP envelope; however, `To` and `Cc` headers are set, while `Bcc` recipients remain hidden (they appear only in the envelope).
*   Internet access to the SMTP server is required.

## Error Handling

The plugin returns descriptive errors for common issues:

*   Invalid or missing configuration fields
*   Malformed JSON in `Params`
*   No recipients specified
*   Failed to decode base64 attachment content
*   SMTP connection or authentication failures
*   Send errors from the SMTP server

All errors are wrapped with context to simplify debugging.

- - -

© 2025 · Built with Go and ❤️