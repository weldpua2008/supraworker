FROM golang:1.12-alpine AS build-env

WORKDIR /go/src/github.com/weldpua2008/supraworker

COPY . .

RUN apk --no-cache add build-base git bzr mercurial gcc
RUN go get -d -v ./... && \
    go install -v ./... && \
    go build -o /root/supraworker main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=build-env /root/supraworker .

CMD ["/root/supraworker","-d"]
