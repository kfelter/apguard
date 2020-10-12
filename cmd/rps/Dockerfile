FROM golang:alpine AS builder
WORKDIR /go/src
ADD . .
RUN cd cmd/rps && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rpx

FROM scratch
COPY --from=builder /go/src/cmd/rps/rpx /rpx
CMD [ "/rpx" ]