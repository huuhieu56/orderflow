FROM golang:1.26-alpine AS build
WORKDIR /src/services/order
COPY libs/ /src/libs/
COPY services/order/ ./
RUN go mod download && go build -o /order ./cmd/order

FROM alpine:3.21
COPY --from=build /order /order
EXPOSE 8080
ENTRYPOINT ["/order"]
