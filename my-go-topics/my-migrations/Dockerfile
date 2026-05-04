# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /src

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy rest
COPY . .

# build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app ./cmd/migrate

# Final stage (minimal image)
FROM scratch
COPY --from=build /app /app

ENTRYPOINT ["/app"]