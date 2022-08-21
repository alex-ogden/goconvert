FROM golang:buster

WORKDIR /usr/src/app

ADD static static
COPY convert.go .

RUN apt-get update
RUN apt-get install -y imagemagick
RUN apt-get install -y libmagickwand-dev
RUN pkg-config --cflags --libs MagickWand
RUN go env -w GO111MODULE=off
RUN go get gopkg.in/gographics/imagick.v2/imagick
RUN export CGO_CFLAGS_ALLOW='-Xpreprocessor'

RUN go build -v -o convert convert.go

CMD ["./convert"]
