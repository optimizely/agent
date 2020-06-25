## Agent Plugins

Plugins allow developers to extend the functionality of Agent through the use of middleware.

### Middleware Plugins

Middleware can be added to agent to customize the request and/or response by implementing the [Middleware](./middleware/registry.go) interface.
This interface needs to define a `Handler()` method that returns a standard net/http middleware handler based on [http.Handler](https://golang.org/pkg/net/http/#Handler) interface.
The custom structure can also define a set of fields that can be configured via `config.yaml`.

### Example Middleware definition
```go
package example

import (
	"context"
	"net/http"

	"github.com/optimizely/agent/plugins/middleware"
)

type MiddlewarePlugin struct {
	// set of configuration fields
	RequestHeader  string
	ResponseHeader string
	ContextValue   string
}

func (p *MiddlewarePlugin) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("X-Example-Request", p.RequestHeader)

			// Example adding context to the request path
			ctx := context.WithValue(r.Context(), "example-context", p.ContextValue)

			// Continuing with the normal serving
			next.ServeHTTP(w, r.WithContext(ctx))

			// Modify the response in some way
			w.Header().Add("X-Example-Response", p.ResponseHeader)
		})
	}
}

// Register our middleware as "example".
func init() {
	middleware.Add("example", func() middleware.Middleware {
		return &MiddlewarePlugin{}
	})
}
```

Once your plugin is written, you'll need to add the plugin as an anonymous import into [all.go](./middleware/all/all.go).
This step is necessary to register the plugin with Agent. 
```go
package all

// Add imports here to trigger the plugin `init()` function
import (
    _ "github.com/optimizely/agent/plugins/middleware/example"
)
```

Example Middleware configuration:
```yaml
server:
  plugins:
    example:
      requestHeader: "example-request"
      responseHeader: "example-response"
      contextValue: "example-context"
```
