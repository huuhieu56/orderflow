FROM golang:1.26-alpine AS build
WORKDIR /src/services/product
COPY libs/ /src/libs/
COPY services/product/ ./
RUN go mod download && go build -o /product ./cmd/product

FROM alpine:3.21
COPY --from=build /product /product
EXPOSE 8080
ENTRYPOINT ["/product"]
