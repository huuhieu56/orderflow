module orderflow/product

go 1.26.5

require (
	github.com/lib/pq v1.12.3
	github.com/redis/go-redis/v9 v9.22.0
	orderflow/platform/httpx v0.0.0
	orderflow/platform/identity v0.0.0
	orderflow/platform/sqlmigrate v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace orderflow/platform/httpx => ../../libs/httpx

replace orderflow/platform/identity => ../../libs/identity

replace orderflow/platform/sqlmigrate => ../../libs/sqlmigrate
