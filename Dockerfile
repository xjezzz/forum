FROM golang:1.20-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash  make gcc gettext musl-dev

COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY ./ ./
RUN go build -o ./bin/app cmd/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY ["/config/db_init.sql", "/config/config.json", "/config/mock_data.sql", "./config/"]
COPY ["/config/certificates/localhost.crt", "/config/certificates/localhost.key", "./config/certificates/"]
COPY web ./web
COPY static ./static
EXPOSE 8080

CMD ["/app"]