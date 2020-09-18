## Agent Plugins

Plugins allow developers to extend the functionality of Agent through the use of middleware.

### Interceptor Plugins

Interceptors can be added to agent to customize the request and/or response by implementing the [Interceptor](./interceptors/registry.go) interface.
This interface needs to define a `Handler()` method that returns a standard net/http middleware handler based on [http.Handler](https://golang.org/pkg/net/http/#Handler).
The custom structure can also include a set of fields that can be configured via `config.yaml`.

### Example Interceptor definition
```go
package example

import (
	"context"
	"net/http"

	"github.com/optimizely/agent/plugins/interceptors"
)

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

To make the interceptor available, add the plugin as an anonymous import into [all.go](./interceptors/all/all.go).
This will register the plugin with Agent. 
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
