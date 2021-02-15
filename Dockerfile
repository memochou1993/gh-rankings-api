# build stage
FROM golang:latest as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/docker.env .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/main .

ENTRYPOINT ./main
