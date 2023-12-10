FROM golang:1.17 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .

# Set environment variables (Replace with your actual values)
ENV AWS_REGION=
ENV AWS_ACCESS_KEY=
ENV AWS_SECRET_KEY=
ENV S3_BUCKET_NAME=

CMD ["./app"]
