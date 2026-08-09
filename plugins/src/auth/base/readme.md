# Auth plugin

Authorize requests by `X-Token` field

## Config section

```yaml
  plugin: ./../plugins/dist/auth/base
  params:
    headerToken: "X-Token" # Header key
    adminToken: ${APP_HTTP_ADMIN_TOKEN} # Admin token - allow all actions
    publicToken: ${APP_HTTP_PUBLIC_TOKEN} # Public token - allow only public actions
```
