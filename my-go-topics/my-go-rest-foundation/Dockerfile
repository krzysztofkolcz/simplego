# FROM golang:1.23-alpine AS build
# WORKDIR /src
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./

# FROM alpine:3.18
# COPY --from=build /app /app
# EXPOSE 8080
# ENV APP_NAME=my-go-rest-foundation
# ENTRYPOINT ["/app"]

# # ---------- BUILD ----------
# FROM golang:1.23-alpine AS builder

# WORKDIR /app

# # cache deps
# COPY go.mod go.sum ./
# RUN go mod download

# # copy source
# COPY . .

# # build binary
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
#     go build -o app ./cmd/server


# # ---------- RUNTIME ----------
# FROM gcr.io/distroless/base-debian12

# WORKDIR /app
# COPY --from=builder /app/app /app/app

# EXPOSE 8080
# USER nonroot:nonroot

# ENTRYPOINT ["/app/app"]



FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/server


FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/app /app/app

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
