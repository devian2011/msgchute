## Architecture and Connecting Message Providers (Plugins)

The notification service utilizes a plugin architecture based on **HashiCorp go-plugin** and **gRPC**. Each provider is compiled into a separate executable binary. The main service runs these binaries in isolated processes and communicates with them via a local gRPC interface.

The interaction is defined by the `provider.proto` contract and the `ProviderService` interface:
1. `Configure(ConfigureRequest) returns (Response)` — passes connection configuration (`params`) on startup.
2. `Send(MessageRequest) returns (MsgResponse)` — triggers a message delivery.

---

### Provider Configuration (`config.yaml`)

Plugins and message delivery workflows are managed under the `providers` configuration block:

```yaml
providers:
  maxBufferSize: 1000       # Maximum queue buffer size for messages
  fetchTaskTimeout: 2s      # Timeout for fetching a task from the queue
  
  # Specification of plugin binary paths (HashiCorp go-plugin)
  pluginMap:
    telegram: ./plugins/dist/providers/tg
    smtp: ./plugins/dist/providers/smtp
    
  # Instance-specific provider settings
  providers:
    gmail:
      provider: smtp        # References a plugin from the pluginMap above
      settings:
        workers:            # Worker pool configuration for concurrent delivery
          min: 2
          max: 6
        breaker:            # Circuit Breaker to prevent error cascading
          windowSize: 10m
          minRequests: 100
          failureThreshold: 0.5 # Disables provider if >50% of requests fail
          timeout: 5m       # Duration the provider remains blocked after a trip
      params:               # Sent to the Configure method as JSON ([]byte)
        host: ://gmail.com
        port: 587
        user: user
        password: password
        from: mail@example.com
```

---

### Creating a New Plugin

#### 1. Initializing HashiCorp go-plugin inside the Provider
On startup, the plugin binary must perform a handshake protocol with the main service. Ensure that your plugin sets up a `plugin.ServeConfig` with configuration keys identical to those in the main service.

Example `main.go` initialization for a custom plugin:
```go
package main

import (
   "github.com/hashicorp/go-plugin"
   "google.golang.org/grpc"
   "github.com/devian2011/msgchute/pkg/shared/auth"
)

type MyProviderServer struct {
	provider.UnimplementedProviderServiceServer
}

// Implement Configure and Send methods here...

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "NOTIFIER_PROVIDER_PLUGIN",
			MagicCookieValue: "stable",
		},
		Plugins: map[string]plugin.Plugin{
			"provider": &ProviderGRPCPlugin{Impl: &MyProviderServer{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
```

#### 2. Handling Parameters (`Configure`)
The values defined in the `params` section of the config file are automatically serialized into JSON by the main service and passed into the `Configure` method as a byte array (`ConfigureRequest.Params`).
Deserialize them inside your plugin as follows:

```go
func (p *MyProviderServer) Configure(ctx context.Context, req *provider.ConfigureRequest) (*provider.Response, error) {
	var cfg MyCustomParams
	if err := json.Unmarshal(req.Params, &cfg); err != nil {
		return &provider.Response{Error: err.Error()}, nil
	}
	// Initialize internal clients/transports here
	return &provider.Response{}, nil
}
```

---

### Error Handling, Retries, and Circuit Breakers

The core service provides built-in mechanisms to ensure fault tolerance when managing plugins:

1. **Retry Policy (Retries)**:
   The service integrates `github.com/devian2011/retrier` to seamlessly recover from transient message delivery errors.
    * If a plugin returns a gRPC-level error or returns an explicit error string within `MsgResponse.error`, the system automatically schedules the message for a retry.
    * **Critical Errors**: Setting `isCritical == true` in the plugin's response signals a fatal, non-recoverable error (e.g., "invalid email format" or "expired API token"). In this scenario, the `retrier` **will bypass** further attempts, marking the message as permanently failed.

2. **Circuit Breaker Protection**:
   If a plugin encounters a spike in failures (crossing the `failureThreshold` within the `windowSize` frame), the Circuit Breaker trips. The main service temporarily stops sending messages to this provider for the duration of the `timeout`. This protects external API rate limits and prevents blocking the shared message processing worker pools.
