FROM golang:buster

WORKDIR /usr/src/app

ADD static static
COPY convert.go .

RUN apt-get update \
    && apt-get install -y imagemagick \
    && apt-get install -y libmagickwand-dev \
    && pkg-config --cflags --libs MagickWand \
    && go env -w GO111MODULE=off \
    && go get gopkg.in/gographics/imagick.v3/imagick \
    && export CGO_CFLAGS_ALLOW='-Xpreprocessor'

RUN go build -v -o convert convert.go

CMD ["./convert"]
