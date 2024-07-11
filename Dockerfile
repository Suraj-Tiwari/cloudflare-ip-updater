FROM golang:1.21.4  as builder
WORKDIR /app
COPY . .
RUN go build -v

FROM alpine:latest
RUN apk add --no-cache libc6-compat
COPY --from=builder /app/cloudflare-ip-updater /usr/bin/
ENTRYPOINT [ "/usr/bin/cloudflare-ip-updater" ]
