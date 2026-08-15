FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /root/

COPY wallet-api .

EXPOSE 8080

CMD ["./wallet-api"]