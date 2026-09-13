# Go Starter

Gin + GORM REST API starter with JWT authentication, refresh tokens, email
verification, password reset, RBAC, Redis caching, Prometheus metrics, tests,
Docker, and GitHub Actions CI.

## Run the project

Copy the example configuration and install Go dependencies:

```sh
cp .env.example .env
go mod tidy
```

Start the API and its dependencies with Docker Compose:

```sh
make docker-up
curl http://localhost:8080/health
```

Useful commands:

```sh
make run       # run the API locally
make build     # build bin/api
make test      # run tests with the race detector
make swagger   # generate OpenAPI files in docs/
make docker-logs
make docker-down
```

`make run` expects PostgreSQL and Redis to be available and the required
variables in `.env` to be set. The default local services are provided by
`make docker-up`.

## Layout

Feature folders in `internal/repository/` and `internal/service/`:

- `internal/repository/user/`         — user repository
- `internal/repository/token/`        — one-time token repository
- `internal/repository/refresh_token/`— refresh token repository
- `internal/repository/rbac/`         — roles & permissions repository
- `internal/service/user/`            — user service
- `internal/service/auth/`            — auth + notifier
- `internal/service/rbac/`            — RBAC service
- `internal/service/admin/`           — admin service
- `internal/service/cleanup/`         — background cleanup

All tests live under `tests/`; the only test helper under `internal/` is
`internal/testutil/db.go`.
