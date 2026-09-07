module orderflow/platform/identity

go 1.26.5

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	orderflow/platform/httpx v0.0.0
)

replace orderflow/platform/httpx => ../httpx
