FROM alpine:3.10
RUN apk update && apk add --no-cache ca-certificates tzdata curl

COPY ./bin/ensemble /usr/bin
COPY ./scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["server"]