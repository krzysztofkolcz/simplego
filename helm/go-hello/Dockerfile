# syntax=docker/dockerfile:1
FROM golang:1.20-alpine AS build
WORKDIR /src
# COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./

FROM alpine:3.18
COPY --from=build /app /app
EXPOSE 8080
ENV APP_NAME=go-hello
ENTRYPOINT ["/app"]