FROM alpine:3.22

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata \\
    && addgroup -S app \\
    && adduser -S -G app app

ADD release /usr/local/src/greeter/

WORKDIR /usr/local/src/greeter
RUN chmod +x ./start.sh ./greeter

USER app
EXPOSE 50051 50052
ENTRYPOINT ["./start.sh"]
