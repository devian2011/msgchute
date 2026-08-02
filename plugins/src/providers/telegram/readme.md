# 📨 Telegram Provider Plugin

A plugin for sending notifications via the Telegram Bot API, designed to be loaded and managed by the **msgchute** plugin system.

## Overview

This plugin implements the `provider.Provider` interface and enables:

*   Text message delivery with HTML or Markdown formatting
*   Attachments (files, images, documents) sent as base64-encoded payloads
*   Broadcast to multiple recipients in one call
*   Silent (notification‑disabled) messages

## Building

Compile the plugin into a binary that can be discovered by the plugin manager:

```
go build -o telegram-provider .
```

## Configuration

The plugin expects a JSON configuration object with the following field:

| Field | Type | Description |
| --- | --- | --- |
| `token` | string | Telegram Bot API token obtained from [@BotFather](https://t.me/botfather) |

Example:

```
{
    "token": "123456:ABC-DEF-ghijklmnopqrstuvwxyz"
}
```

## Usage

The plugin exposes the `Send(msg *provider.Message) error` method. The `Message` structure is defined as:

```
type Message struct {
    To      []string // list of chat_id (as strings)
    Subject string   // subject line (displayed in bold)
    Body    string   // message body
    Params  []byte   // JSON-encoded SendOptions (see below)
}
```

### Send Options (`Params`)

The `Params` field must contain a JSON object with these optional properties:

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `parse_mode` | string | `"HTML"` | Either `"HTML"` or `"Markdown"` |
| `disable_notification` | boolean | `false` | If `true`, sends the message silently |
| `attachments` | array | —   | List of files to attach (see below) |

#### Attachment Object

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `filename` | string | ✅   | Name of the file |
| `content` | string | ✅   | Base64‑encoded file content |
| `content_type` | string | ❌   | MIME type (defaults to `application/octet-stream`) |

## Examples

### 1\. Simple text message

```
msg := &provider.Message{
    To:      []string{"123456789"},
    Subject: "Hello",
    Body:    "This is a test message.",
    Params:  []byte(`{"parse_mode":"HTML"}`),
}
```

### 2\. Message with an attachment

```
import "encoding/base64"

fileContent := base64.StdEncoding.EncodeToString([]byte("Hello, world!"))
params := map[string]interface{}{
    "parse_mode": "Markdown",
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
    To:      []string{"123456789"},
    Subject: "File attached",
    Body:    "Here is your file:",
    Params:  paramsJSON,
}
```

### 3\. Broadcast to multiple recipients

```
msg := &provider.Message{
    To:      []string{"123456789", "987654321"},
    Subject: "Announcement",
    Body:    "This message is sent to everyone.",
    Params:  []byte(`{}`),
}
```

## Notes

*   `chat_id` values are strings but internally converted to `int64`. For group chats, use a negative ID (e.g., `"-100123456789"`).
*   When attachments are present, the plugin sends the text message first, followed by each attachment as a separate document. The order is not guaranteed but is sequential per recipient.
*   All attachments are sent using the `sendDocument` method; for other types (photos, audio, etc.) you can set the appropriate `content_type`, but the actual method remains `sendDocument`.
*   Internet access to the Telegram API is required.

## Error Handling

The plugin returns descriptive errors for common issues:

*   Invalid or missing token
*   Malformed JSON in `Params`
*   Invalid `chat_id` (non‑numeric)
*   Failed to decode base64 attachment content
*   Telegram API errors (e.g., bot blocked, chat not found)

All errors are wrapped with context to simplify debugging.

- - -

© 2025 · Built with Go and ❤️