module github.com/optimizely/agent

go 1.13

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/httplog v0.1.6
	github.com/go-chi/render v1.0.1
	github.com/go-kit/kit v0.9.0
	github.com/google/uuid v1.1.1
	github.com/lestrrat-go/jwx v0.9.0
	github.com/optimizely/go-sdk v1.5.0
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.18.1-0.20200514152719-663cbb4c8469
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	gopkg.in/yaml.v2 v2.2.4
)

exclude github.com/gorilla/websocket v1.4.0

replace github.com/coreos/etcd v3.3.10+incompatible => github.com/etcd-io/etcd v3.3.25+incompatible
