# Dependency security notes

This document records decisions for Arnica (and similar SCA/supply-chain)
findings that are **accepted and documented** rather than remediated by a
version bump or replacement. Findings that were remediated (vulnerable
dependency upgrades, the Go toolchain bump, the notification handler change,
and the container healthchecks) are not listed here — they live in the code and
`go.mod`/requirements history.

Reviewed against the repository's Apache-2.0 license.

## Reputation: low-reputation dependencies (accepted)

All five flagged modules are **indirect** dependencies. They are not imported
directly by Agent code; each is pulled in by a well-maintained, widely used
direct dependency. They cannot be replaced without replacing their parent, and
no better-maintained drop-in exists. We accept them.

| Module | Pulled in by | Used for |
| --- | --- | --- |
| `github.com/lestrrat-go/httpcc` | `github.com/lestrrat-go/jwx/v2/jwk` | JWT/JWK verification (`pkg/middleware` auth) |
| `github.com/lestrrat-go/option` | `github.com/lestrrat-go/jwx/v2/jwk` | JWT/JWK verification (`pkg/middleware` auth) |
| `github.com/lestrrat-go/iter` | `github.com/lestrrat-go/jwx/v2/jwk` | JWT/JWK verification (`pkg/middleware` auth) |
| `github.com/VividCortex/gohistogram` | `github.com/go-kit/kit/metrics` | expvar/generic metrics (`pkg/metrics`) |
| `github.com/matttproud/golang_protobuf_extensions/v2` | `github.com/prometheus/client_golang` | Prometheus metrics (`pkg/metrics`) |

Notes:

- The `lestrrat-go/*` helpers are maintained as part of the `jwx` project, which
  is the JWT library this service depends on for request authentication.
- `matttproud/golang_protobuf_extensions/v2` was mis-listed as `mattproud`
  (single `t`) in the Arnica list view; the real module name has three `t`s. It
  is an implementation detail of the Prometheus client and is only reachable
  through it.

## License: MPL-2.0 dependencies (accepted)

MPL-2.0 is a weak, file-level reciprocal license. Its source-disclosure
obligation is triggered only when the MPL-licensed **files themselves** are
modified and distributed. Agent consumes all of these as unmodified upstream
dependencies, so no reciprocal obligation is triggered, and MPL-2.0 is
compatible with this Apache-2.0 project. We accept them.

| Dependency | Scope | Pulled in by |
| --- | --- | --- |
| `github.com/hashicorp/hcl@v1.0.0` | Go, indirect | `github.com/spf13/viper` (config parsing) |
| `github.com/hashicorp/go-multierror@v1.1.1` | Go, indirect | `github.com/optimizely/go-sdk/v2` client |
| `github.com/hashicorp/errwrap@v1.1.0` | Go, indirect | `github.com/hashicorp/go-multierror` |
| `certifi` | Python, `examples/requirements.txt` | example SSE client script only |

Notes:

- The `hashicorp/*` modules are unmodified, standard transitive dependencies and
  are not vendored or patched.
- `certifi` ships only in the Python **example** requirements (not in the Agent
  binary or any released artifact). It was also upgraded to the latest release
  to clear its vulnerability finding.

## Container image healthchecks

- `scripts/dockerfiles/Dockerfile.alpine` defines a `HEALTHCHECK` that probes the
  admin listener's `/health` endpoint using the `wget` bundled in the Alpine base
  image.
- `scripts/dockerfiles/Dockerfile.static` is a `FROM scratch` image with no shell
  or HTTP client to run a `HEALTHCHECK` command, and the binary exposes no health
  subcommand. Health must be checked over HTTP against the `/health` endpoint via
  the orchestrator (e.g. a Kubernetes liveness/readiness probe). This is
  documented inline in that Dockerfile.
