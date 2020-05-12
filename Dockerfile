FROM golang:1.14.2-alpine3.11 as base
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN apk add --no-cache --virtual ca-certificates git make gcc musl-dev
RUN go mod download

FROM base AS builder
COPY . .
RUN make test vet
RUN make build

FROM alpine:3.11
WORKDIR /var/lib/archivematica
COPY --from=builder /src/rdss-archivematica-channel-adapter .
RUN apk --no-cache add ca-certificates
RUN addgroup -g 333 -S archivematica && adduser -u 333 -h /var/lib/archivematica -S -G archivematica archivematica
USER archivematica
ENTRYPOINT ["/var/lib/archivematica/rdss-archivematica-channel-adapter"]
CMD ["help"]
