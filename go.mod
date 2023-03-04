module github.com/optimizely/agent

go 1.20

require (
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/cors v1.2.1
	github.com/go-chi/httplog v0.2.5
	github.com/go-chi/render v1.0.2
	github.com/go-kit/kit v0.12.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/lestrrat-go/jwx v0.9.0
	github.com/optimizely/go-sdk v1.8.4-0.20230216074708-27b2772ccf33
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.27.0
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.0.0-20210915214749-c084706c2272
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-chi/chi/v5 v5.0.8 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/twmb/murmur3 v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude github.com/gorilla/websocket v1.4.0

replace github.com/coreos/etcd v3.3.10+incompatible => github.com/etcd-io/etcd v3.3.25+incompatible
