# syntax=docker/dockerfile:1


## Build 
FROM golang:1.21 AS Build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o orders-manager .

EXPOSE 8080

CMD [ "./orders-manager", "start" ]