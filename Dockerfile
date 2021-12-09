FROM golang:1.17.5
LABEL maintainer="Transnano <transnano.jp@gmail.com>"
WORKDIR /go/src
ENV GO111MODULE=on
RUN go mod download
EXPOSE 8080
CMD ["go", "run", "/go/src/main.go"]
