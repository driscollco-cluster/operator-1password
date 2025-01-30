FROM alpine:latest
RUN apk update && apk add ca-certificates && apk add tzdata && apk add curl && rm -fr /var/cache/apk*
ADD config.yml /config.yml
ADD service /service
ENTRYPOINT ["/service"]