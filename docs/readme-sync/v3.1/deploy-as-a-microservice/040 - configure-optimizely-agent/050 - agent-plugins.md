---
title: "Agent plugins"
excerpt: ""
slug: "agent-plugins"
hidden: false
metadata:
  title: "Agent plugins - Optimizely Full Stack"
createdAt: "2020-09-21T20:30:00.000Z"
updatedAt: "2020-09-21T20:30:00.000Z"
---

## Agent Plugins

Optimizely Agent can be extended through the use of plugins. Plugins are distinct from the standard Agent packages that provide a namespaced environment for custom logic. Plugins must be compiled as part of the Agent distribution and are enabled through configuration.

### Interceptor Plugins

Interceptors can be added to Agent to customize the request and/or response by implementing the [Interceptor](https://github.com/optimizely/agent/tree/master/plugins/interceptors/registry.go) interface.
This interface defines a `Handler()` method that returns a standard net/http middleware handler based on [http.Handler](https://golang.org/pkg/net/http/#Handler).
The interceptor struct can also include a set of fields that can be configured via `config.yaml`.

* [httplog](https://github.com/optimizely/agent/tree/master/plugins/interceptors/httplog) - Adds HTTP request logging based on [go-chi/httplog](https://github.com/go-chi/httplog).

### Example Interceptor definition
```go
package example

import (
	"context"
	"net/http"

	"github.com/optimizely/agent/plugins/interceptors"
)

// Example implements the Interceptor plugin interface
type Example struct {
	// set of configuration fields
	RequestHeader  string
	ResponseHeader string
	ContextValue   string
}

func (i *Example) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("X-Example-Request", i.RequestHeader)

			// Example adding context to the request path
			ctx := context.WithValue(r.Context(), "example-context", i.ContextValue)

			// Continuing with the normal serving
			next.ServeHTTP(w, r.WithContext(ctx))

			// Modify the response in some way
			w.Header().Add("X-Example-Response", i.ResponseHeader)
		})
	}
}

// Register our interceptor as "example".
func init() {
	interceptors.Add("example", func() interceptors.Interceptor {
		return &Example{}
	})
}
```

To make the interceptor available to Agent, add the plugin as an anonymous import into [all.go](./interceptors/all/all.go).
```go
package all

// Add imports here to trigger the plugin `init()` function
import (
    _ "github.com/optimizely/agent/plugins/interceptors/example"
)
```

Enable the example interceptor by adding to `server.interceptors` within your `config.yaml`. Note that the yaml fields should match the struct definition of your plugin.
```yaml
server:
  interceptors:
    example:
      requestHeader: "example-request"
      responseHeader: "example-response"
      contextValue: "example-context"
```
