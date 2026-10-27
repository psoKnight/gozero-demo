FROM alpine:3.22

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata \\
    && addgroup -S app \\
    && adduser -S -G app app

ADD release /usr/local/src/calculator/

WORKDIR /usr/local/src/calculator
RUN chmod +x ./start.sh ./calculator

USER app
EXPOSE 50053 50054
ENTRYPOINT ["./start.sh"]
