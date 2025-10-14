FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build both binaries
RUN go build -o bin/api cmd/api/main.go
RUN go build -o bin/worker cmd/worker/main.go

FROM alpine:3.22 AS prod

WORKDIR /app

# Copy both binaries from build stage
COPY --from=build /app/bin/api /app/api
COPY --from=build /app/bin/worker /app/worker

EXPOSE 8080

# Default to API server (can be overridden in docker-compose)
CMD ["./api"]


