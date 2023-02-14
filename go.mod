module github.com/optimizely/agent

go 1.13

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/httplog v0.1.6
	github.com/go-chi/render v1.0.1
	github.com/go-kit/kit v0.9.0
	github.com/go-redis/redis/v8 v8.11.4
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/google/uuid v1.1.1
	github.com/lestrrat-go/jwx v0.9.0
	github.com/optimizely/go-sdk v1.8.4-0.20230117172806-a96c3071c921
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.18.1-0.20200514152719-663cbb4c8469
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/yaml.v2 v2.4.0
)

exclude github.com/gorilla/websocket v1.4.0

replace github.com/coreos/etcd v3.3.10+incompatible => github.com/etcd-io/etcd v3.3.25+incompatible
