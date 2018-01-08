FROM golang:1.9.1-alpine
WORKDIR /go/src/github.com/JiscRDSS/rdss-archivematica-channel-adapter
COPY . .
# Don't use `make testrace`, it won't work in Alpine Linux!
RUN set -x \
	&& apk add --no-cache ca-certificates \
	&& apk add --no-cache --virtual .build-deps make gcc musl-dev git \
	&& make test vet \
	&& make install
RUN set -x \
	&& addgroup -g 333 -S archivematica \
	&& adduser -u 333 -h /var/lib/archivematica -S -G archivematica archivematica
USER archivematica
ENTRYPOINT ["/go/bin/rdss-archivematica-channel-adapter"]
CMD ["help"]
