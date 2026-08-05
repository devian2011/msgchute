## Custom Authorization Extensions (Auth Plugins)

The system supports pluggable API authentication via **HashiCorp go-plugin** and **gRPC**. Instead of making network calls to a remote server, the service spawns the authentication engine as an isolated sub-process from a local binary path and communicates via an internal gRPC channel using the `AuthProviderService` contract.

---

### Configuration

To enable and map your custom authorization engine, configure the `auth` block in your `config.yaml` with the absolute or relative path to your compiled authentication binary:

```yaml
auth:
  plugin: ./plugins/dist/auth/custom-rbac
```

---

### Implementing an Auth Plugin

The custom binary must act as a HashiCorp plugin server, executing a handshake protocol and serving the `AuthProviderService` over gRPC.

#### 1. Setup Plugin Handshake & Server (`main.go`)

Ensure that your `MagicCookieKey` and `MagicCookieValue` match the main service's expectations for auth plugins (e.g., `NOTIFIER_AUTH_PLUGIN`).

```go
package main

import (
	"context"
	
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"github.com/devian2011/msgchute/pkg/shared/auth"
)

type CustomAuthPlugin struct {
	auth.UnimplementedAuthProviderServiceServer
}

// Allow executes the authorization check for every incoming API request
func (s *CustomAuthPlugin) Allow(ctx context.Context, req *auth.AllowRequest) (*auth.AllowResponse, error) {
	// Extract Authorization header values safely
	if authHeader, ok := req.Headers["authorization"]; ok && len(authHeader.Values) > 0 {
		token := authHeader.Values[0]
		
		// Implement your validation token or session check here
		if token == "valid-secret-token" {
			return &auth.AllowResponse{Allowed: true}, nil
		}
	}
	
	// Reject by default if criteria are not met
	return &auth.AllowResponse{Allowed: false}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "NOTIFIER_AUTH_PLUGIN",
			MagicCookieValue: "stable",
		},
		Plugins: map[string]plugin.Plugin{
			"auth_provider": &AuthGRPCPlugin{Impl: &CustomAuthPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
```

#### 2. Request Payload Reference (`AllowRequest`)

When the main gateway routes an HTTP/gRPC request through your auth plugin, it maps the request lifecycle context onto the following gRPC message properties:

* **`Method`**: The request's HTTP method context (e.g., `POST`, `GET`, `DELETE`).
* **`Path`**: The exact routing endpoint request pathname (e.g., `/api/v1/send`).
* **`Headers`**: Map of arrays containing incoming transfer headers. Metadata keys are normalized to lowercase.
* **`Cookies`**: Extracted cookie state map string-pairs for validating browser-based or stateful client sessions.
