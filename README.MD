# MSGChute Notification Service

A high-performance, fault-tolerant microservice written in **Go**, designed around a **push model** to aggregate, route, and deliver messages to external systems and communication channels (such as Email, Telegram, and others).

Built from the ground up for reliable external integration, the service offers highly flexible options for sending both structured, dynamic templates and direct, raw notifications, backed by complete operational visibility.

### Key Features & Capabilities

*   **Push-Driven Architecture**: The service ingests incoming requests through its API, queues them internally, and immediately pushes payloads out to target external systems via managed, concurrent worker pools.
*   **Flexible Templating via `flosch/pongo2`**:
  *   **Template-Based Delivery**: Features a built-in template engine powered by **`flosch/pongo2`** (a Django-like template syntax for Go). You can pass a template identifier along with a context map of dynamic variables to render personalized, rich content on the fly.
  *   **Direct Delivery (Template-less)**: For simple text, alerts, or payloads pre-rendered on the client side, the service fully supports direct sending. You can bypass the template engine entirely by passing a raw `subject` and `body` directly in your request.
*   **Administrative Management API**: Provides a secure suite of administrative endpoints dedicated to operational control and manual intervention:
  *   **Status Monitoring**: Inspect the real-time delivery state, history, and current queues of all outbound messages.
  *   **Manual Retries**: Trigger immediate, manual re-delivery attempts for messages that failed due to transient issues or exhausted automated policies.
  *   **Message Revocation**: Manually cancel or reject pending messages sitting in queues before they are dispatched to external endpoints.
*   **Downstream Response Logging**: Maintains an exhaustive execution audit log. Every single response, error message, status code, and latency metric returned by external services and provider plugins is systematically logged to assist in debugging and service-level tracking.
*   **Extensibility via Binary Plugins**: New communication channels or custom external integrations are added as isolated pluggable extensions powered by **HashiCorp go-plugin** and **gRPC**. Each provider executes as an independent sub-process, ensuring core system stability even if a specific plugin fails.
*   **Adaptive Worker Pools**: Every external provider instance can be fine-tuned with individual concurrency limits (`min`/`max` workers). This allows safe throttling to strictly comply with target API rate limits (e.g., Telegram Bot API message-per-second constraints).
*   **Smart Retries**: Integrated with the `retrier` library to automatically handle transient network glitches or temporary external outages. It respects an `isCritical` flag; if an external system returns a fatal error status (e.g., "invalid credentials" or "recipient blocked"), the service terminates retries immediately to preserve resources.
*   **Circuit Breaker Protection**: Safeguards the ecosystem against cascading downstream failures. If an external API experiences a prolonged outage and its failure rate crosses your configured threshold, the circuit breaker trips, temporarily isolating that provider and protecting internal buffers from filling up.
*   **Custom Authorization Middleware**: Supports slotting in independent authorization binary plugins. The service inspects routing paths, HTTP methods, headers, and cookies to validate inbound push requests before accepting tasks into the queue.
*   **12-Factor App Configuration**: Driven by `gookit/config` with native support for default values, optional `.env` file sourcing during development, and environment variable overrides for production credentials.

## Documentation & Reference

Explore the detailed guides, API contracts, and extension specifications:

* [Configuration](./docs/config.md)
* [Template engine](https://github.com/flosch/pongo2)
* [API](./docs/api.md)
  * [Swagger](./app/docs/swagger.json)
* Extensions
  * [Make own provider](./docs/provider.md)
  * [Make own auth middleware](./docs/auth.md)

