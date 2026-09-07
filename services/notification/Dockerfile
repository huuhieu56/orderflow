FROM golang:1.26-alpine AS build
WORKDIR /src/services/notification
COPY libs/ /src/libs/
COPY services/notification/ ./
RUN go mod download && go build -o /notification ./cmd/notification

FROM alpine:3.21
COPY --from=build /notification /notification
EXPOSE 8080
ENTRYPOINT ["/notification"]
