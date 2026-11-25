FROM golang:1.27.4-alpine AS go-base
# Статические бинарники могут запускаться в минимальных образах на базе scratch.
ENV CGO_ENABLED=0
WORKDIR /src

FROM go-base AS app-build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN go build -trimpath -o /out/pickme ./cmd/pickme

FROM go-base AS migrate-build
# Для миграций PickMe достаточно драйвера PostgreSQL.
RUN go install -tags='no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb' \
    github.com/pressly/goose/v3/cmd/goose@v3.30.1

FROM scratch AS app
COPY --from=go-base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=app-build /out/pickme /pickme
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/pickme"]

FROM scratch AS migrate
COPY --from=go-base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=migrate-build /go/bin/goose /goose
COPY migrations /migrations
WORKDIR /migrations
USER 65532:65532
ENTRYPOINT ["/goose"]
CMD ["up"]
