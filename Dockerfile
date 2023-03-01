# Build stage
FROM golang:alpine AS build

RUN apk update && apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

# Final stage
FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /build/app ./
CMD ["./app"]
