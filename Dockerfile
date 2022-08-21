FROM golang:buster

WORKDIR /usr/src/app

ADD static static
COPY convert.go .

RUN apt-get update \
    && apt-get install -y imagemagick \
    && apt-get install -y libmagickwand-dev \
    && export CGO_CFLAGS_ALLOW='-Xpreprocessor' \
    && pkg-config --cflags --libs MagickWand \
    && go env -w GO111MODULE=off \
    && go clean -cache \
    && go get gopkg.in/gographics/imagick.v3/imagick

RUN go build -v -o convert convert.go

CMD ["./convert"]
