# AGENTS.md

## Architecture

- Microservices inside one git repo. Each service in `apps/` is an independent microservice: own `go.mod`, own database, own deployment, communicates only through contracts. Never treat services as one shared codebase.
- Workspace (`go.work`) + moon only for local dev tooling; services stay independently buildable and deployable.
- Hexagonal: `internal/port` (interfaces, domain), `internal/usecase` (business logic), `internal/adapter` (outbound impl), `internal/interface` (inbound transport).
- Use cases must not import transport or DB packages. Adapters implement ports.
- Shared contracts only: `common-rpc/proto/**` (+ generated `gen/`) via Buf. Business models never live there.
- Services never query another service's database.

## Auth model

- `edge-api` is the only public HTTP gateway. It has no Kratos/Hydra logic.
- `identity-service` (private, gRPC :9081, mTLS) owns Kratos orchestration and issues Ed25519-signed, target-bound, 60s actor assertions (`aud=<service>`).
- `order-service` (private, gRPC :9082, mTLS) verifies caller cert (`edge-api.internal`), assertion signature/iss/aud/exp, then exposes actor via context.
- Never forward raw Kratos session tokens past Identity. Never trust `x-user-id`-style headers. Private networking is not authentication.
- All internal gRPC requires TLS 1.3 mTLS + per-RPC caller ACL.

## Testing

- Unit tests live next to code: `internal/**/**_test.go` (fakes allowed).
- Integration tests cross real boundaries: `apps/*/tests/integration/**`.
- Tests require generated local certs: `./deployments/compose/certs/generate.sh`.

## Commands

```sh
cd common-rpc && buf generate && buf lint && go test ./...
cd apps/<service> && go test -race ./... && go vet ./...
```

## Conventions

- Commit conventional commits: `feat:`, `fix:`, `refactor:`, `chore:`.
- Generated protobuf (`*.pb.go`) is committed, never hand-edited.
- Local secrets/certs ignored: `properties.yaml`, `deployments/compose/certs/*.{crt,key,pub,srl}`.
- Formatting: `gofmt` + blank lines between logical blocks; one declaration/statement per line.
