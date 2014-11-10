FROM golang:1.3.3-onbuild

EXPOSE 4001

CMD /go/bin/km -config=/config/config.yml -workdir=/go/src/github.com/FreekKalter/km
