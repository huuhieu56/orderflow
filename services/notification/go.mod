module orderflow/notification

go 1.26.5

require (
	github.com/lib/pq v1.12.3
	github.com/segmentio/kafka-go v0.4.51
	orderflow/platform/httpx v0.0.0
	orderflow/platform/identity v0.0.0
	orderflow/platform/sqlmigrate v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace orderflow/platform/httpx => ../../libs/httpx

replace orderflow/platform/identity => ../../libs/identity

replace orderflow/platform/sqlmigrate => ../../libs/sqlmigrate
