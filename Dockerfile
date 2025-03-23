FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app
RUN adduser -D -g '' app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o build/echoy .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /app/build/echoy /usr/local/bin/echoy

USER app

ENV TZ=UTC \
    APP_USER=app

ENTRYPOINT ["/usr/local/bin/echoy"]

CMD []