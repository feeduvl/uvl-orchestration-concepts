FROM golang:1.19.4
WORKDIR /go/src/app
COPY . .
RUN go install -d -v ./...
RUN go install -v ./...

EXPOSE 9709
CMD ["app"]
