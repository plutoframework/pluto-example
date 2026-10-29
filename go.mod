module github.com/plutoframework/pluto-example

go 1.25.5

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/jackc/pgx/v5 v5.11.0
	github.com/plutoframework/pluto-auth v0.3.0
	github.com/plutoframework/pluto-auth/methods/password v0.3.0
	github.com/plutoframework/pluto-auth/stores/postgres v0.3.0
	github.com/plutoframework/pluto-core v0.3.0
	github.com/plutoframework/pluto-core/stores/postgres v0.1.0
	github.com/plutoframework/pluto-core/stores/valkey v0.1.0
	github.com/plutoframework/pluto-jobs v0.1.1
	github.com/plutoframework/pluto-jobs/stores/postgres v0.1.0
	github.com/plutoframework/pluto-rbac v0.1.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/valkey-io/valkey-go v1.0.79 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
