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

## Tooling

- Go toolchain is enabled in `.moon/toolchains.yml` (workspaces on, `go.work` aware).
- `buf`, `protoc-gen-go`, `protoc-gen-go-grpc` are pinned via the go toolchain `bins` and auto-installed.
- No pinned Go version = uses system `go`; pin `version` or add `.prototools` for determinism.
