FROM golang:1.17

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o trader ./cmd/tradebot

CMD ["/app/trader"]