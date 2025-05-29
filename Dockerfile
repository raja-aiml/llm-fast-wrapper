FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
 RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o llm-fast-wrapper ./server
 # Build CLI binaries for database migrations and intent seeding
 RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrate-cli ./cmd/migrate
 RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o intent-cli ./cmd/intent

FROM gcr.io/distroless/static-debian11
WORKDIR /app
 COPY --from=builder /app/llm-fast-wrapper .
 # Include CLI executables for ArgoCD jobs
 COPY --from=builder /app/migrate-cli .
 COPY --from=builder /app/intent-cli .
ENTRYPOINT ["/app/llm-fast-wrapper"]
CMD ["serve", "--fiber"]
