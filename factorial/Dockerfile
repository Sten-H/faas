FROM golang:1.9

RUN go get -u github.com/kardianos/govendor

COPY . /go/src/github.com/sten-H/faas/factorial
WORKDIR /go/src/github.com/sten-H/faas/factorial

# fetch missing dependencies
CMD govendor fetch +m

RUN go get ./
RUN go build

# will use pilu/fresh for code reloading via docker-compose volume sharing with local machine
CMD go get github.com/pilu/fresh && fresh

#EXPOSE 8080
