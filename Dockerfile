FROM golang:latest

WORKDIR /usr/src/app
ADD static static
COPY convert.go .
RUN go build -v -o convert convert.go

CMD ["./convert"]
