FROM golang:1.15

WORKDIR /go/src/github.com/go-mysql

COPY . .

RUN go get -u github.com/go-sql-driver/mysql

RUN go build -o main

CMD ["./main"]