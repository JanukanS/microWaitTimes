# syntax=docker/dockerfile:1

FROM golang:1.16-bullseye
WORKDIR /app

COPY src/* ./
RUN go mod download
RUN go build -o /ms-wait-times
CMD [ "/ms-wait-times" ]