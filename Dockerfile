FROM golang:1.19.4
WORKDIR /go/src/app
COPY . .
RUN go install -d -v ./...

EXPOSE 9709
CMD ["app"]
