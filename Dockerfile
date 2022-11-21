FROM golang:1.19.2-alpine

MAINTAINER awohsen <awohsen@gmail.com>

WORKDIR /app
COPY . .

RUN go get -d -v
RUN go build -v

CMD ["./telereq"]