FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
WORKDIR /root/

RUN apk --no-cache add ca-certificates postgresql-client mysql-client 

COPY --from=builder /app/main /bin/postgres-s3-backup

CMD ["/bin/postgres-s3-backup"]
