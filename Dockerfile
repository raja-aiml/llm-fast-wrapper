FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o llm-fast-wrapper ./cmd/main.go
FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/llm-fast-wrapper .
ENTRYPOINT ["/app/llm-fast-wrapper"]
CMD ["serve", "--fiber"]
