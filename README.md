# E-Commerce GO

Microservices monorepo (Go) orchestrated with [moon](https://moonrepo.dev).

## Structure

```
apps/                     # deployable services (each own go.mod)
  edge-api/               # gateway (gin)
  catalog-service/        # :8081
  order-service/          # :8082
  inventory-service/      # :8083
  payment-service/        # :8084
  notification-service/   # :8085
common-rpc/               # shared gRPC protos + generated code (buf)
common-events/            # shared event protos + generated code (buf)
platform/observability/   # shared logging helpers
go.work                  # Go workspace (all modules)
```

## Commands

```sh
moon query projects        # list all projects
moon run :build            # build all modules
moon run :test             # test all modules
moon run :lint             # go vet all modules
moon run :fmt              # check formatting
moon run common-rpc:generate   # regenerate protos
moon run common-rpc:protolint  # buf lint
moon run common-rpc:breaking   # buf breaking vs main
moon run apps/edge-api:run     # run a service locally (PORT=8080..8085)
moon run apps/edge-api:dev     # run with air hot-reload (dev)
moon ci                    # all checks (CI entry point)
```

## Secure auth example: Edge → Identity → Order

The protected `GET /auth/order-caller` route proves the complete trust chain:

1. Edge calls private `identity-service.ResolveSession` using TLS 1.3 mTLS.
2. Identity resolves the Kratos session and signs a 60-second Ed25519 actor assertion for `order-service`.
3. Edge calls Order using a separate mTLS connection and forwards only that assertion.
4. Order accepts only `edge-api.internal`, verifies assertion signature/issuer/audience/expiry, then exposes actor context to its use case.

```sh
# one-time local development keys and workload certificates (ignored by git)
./deployments/compose/certs/generate.sh
cp apps/identity-service/properties.example.yaml apps/identity-service/properties.yaml

# Ory must already be running
# terminal 1
cd apps/identity-service && go run ./cmd/identity-service

# terminal 2
cd apps/order-service && go run ./cmd/order-service

# terminal 3
cd apps/edge-api && go run ./cmd/edge-api

# use the session_token returned by POST /auth/login
curl -H 'Authorization: Bearer <session_token>' \
  http://localhost:8080/auth/order-caller
```

Ports/adapters: `common-rpc/proto/{identity,order}/v1`, `apps/identity-service/internal`, `apps/edge-api/internal/{port,adapter}`, and `apps/order-service/internal`.

## Tooling

- Go toolchain is enabled in `.moon/toolchains.yml` (workspaces on, `go.work` aware).
- `buf`, `protoc-gen-go`, `protoc-gen-go-grpc` are pinned via the go toolchain `bins` and auto-installed.
- No pinned Go version = uses system `go`; pin `version` or add `.prototools` for determinism.
