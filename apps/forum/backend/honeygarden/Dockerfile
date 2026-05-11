FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /honeygarden ./cmd/main.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /honeygarden /honeygarden
EXPOSE 8080
ENTRYPOINT ["/honeygarden"]
