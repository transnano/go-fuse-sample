FROM golang:1.15.2
LABEL maintainer "[my name]"
WORKDIR /go/src
ENV GO111MODULE=on
RUN go mod download
EXPOSE 8080
CMD ["go", "run", "/go/src/main.go"]
