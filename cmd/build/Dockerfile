FROM alpine:latest
RUN apk update && apk add ca-certificates && apk add tzdata && apk add curl && rm -fr /var/cache/apk*
ADD service /service
ENTRYPOINT ["/service"]