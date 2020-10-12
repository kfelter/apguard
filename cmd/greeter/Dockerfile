FROM golang:alpine AS builder
WORKDIR /go/src
ADD . .
RUN cd cmd/greeter && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o g

FROM scratch
COPY --from=builder /go/src/cmd/greeter/g /g
CMD [ "/g" ]