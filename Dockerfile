FROM golang:alpine

WORKDIR /usr/src/app

ADD ["static", "static"]
ADD ["convert", "convert"]

RUN apk --update add build-base imagemagick imagemagick-dev git \
    && export CGO_CFLAGS_ALLOW='-Xpreprocessor' \
    && pkg-config --cflags --libs MagickWand \
    && go mod init convert \
    && go get gopkg.in/gographics/imagick.v3/imagick

WORKDIR /usr/src/app/convert
RUN go build

CMD ["./convert"]
