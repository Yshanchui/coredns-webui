FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o coredns-webui .

FROM alpine:latest
WORKDIR /app
# Copy CoreDNS binary for validation
COPY --from=coredns/coredns:latest /coredns /usr/local/bin/coredns
COPY --from=builder /app/coredns-webui .
# Copied templates are embedded now, but good practice to keep structure if needed.
# Since we embedded them with //go:embed, strictly we just need the binary.

EXPOSE 80
CMD ["./coredns-webui"]
