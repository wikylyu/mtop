FROM alpine:latest

WORKDIR /app
COPY ./bin ./bin
COPY ./keys ./keys
COPY ./climber.yaml .
COPY ./mtop.yaml .

EXPOSE 10811
EXPOSE 10801

CMD ["./bin/climber"]