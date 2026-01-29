FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./

FROM alpine:3.18
COPY --from=build /app /app
EXPOSE 8080
ENV APP_NAME=my-go-rest-foundation
ENTRYPOINT ["/app"]