FROM golang:1.24-alpine AS build
RUN --mount=type=cache,target=/var/lib/apk \
    apk add tini-static sqlite-dev build-base

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -o /moneyprinter2 -ldflags '-extldflags "-static"' .

FROM scratch
COPY --from=build /sbin/tini-static /tini
COPY --from=build /moneyprinter2 /moneyprinter2
ENV MONEYPRINTER_DB=/data/mp2.db
ENTRYPOINT ["/tini", "/moneyprinter2", "serve"]
