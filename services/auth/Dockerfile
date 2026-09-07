FROM golang:1.26-alpine AS build
WORKDIR /src/services/auth
COPY libs/ /src/libs/
COPY services/auth/ ./
RUN go mod download && go build -o /auth ./cmd/auth

FROM alpine:3.21
COPY --from=build /auth /auth
EXPOSE 8080
ENTRYPOINT ["/auth"]
