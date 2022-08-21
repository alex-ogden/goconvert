FROM golang:alpine

WORKDIR /usr/src/app

ADD static static
COPY convert.go .

RUN apk --update add build-base imagemagick imagemagick-dev git \
    && export CGO_CFLAGS_ALLOW='-Xpreprocessor' \
    && pkg-config --cflags --libs MagickWand \
    && go env -w GO111MODULE=off \
    && go get gopkg.in/gographics/imagick.v3/imagick

RUN go build -v -o convert convert.go

CMD ["./convert"]
