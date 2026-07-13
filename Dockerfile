FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /goshrt ./cmd/goshrt

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /goshrt /usr/local/bin/goshrt

EXPOSE 8080
CMD ["goshrt"]
