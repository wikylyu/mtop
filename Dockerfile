FROM alpine:latest

WORKDIR /app
COPY ./bin ./bin
COPY ./keys ./keys
COPY ./climber.yaml .
COPY ./mtop.yaml .

EXPOSE 1081
EXPOSE 1080

CMD ["./bin/climber"]