FROM golang:1.19.4
WORKDIR /go/src/app
COPY . .
RUN go mod ./...
RUN go install -v ./...

EXPOSE 9709
CMD ["app"]
