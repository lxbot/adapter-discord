FROM golang:1.13.6-alpine3.11 as builder

ARG GOLANG_NAMESPACE="github.com/lxbot/adapter-discord"
ENV GOLANG_NAMESPACE="$GOLANG_NAMESPACE"

RUN apk --no-cache add alpine-sdk coreutils make tzdata
RUN cp -f /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
WORKDIR /go/src/$GOLANG_NAMESPACE
ADD ./go.* /go/src/$GOLANG_NAMESPACE/
ENV GO111MODULE=on
RUN go mod download
ADD . /go/src/$GOLANG_NAMESPACE/
RUN make build
RUN mkdir -p /lxbot/adapters
RUN mv /go/src/$GOLANG_NAMESPACE/adapter-discord.so /lxbot/adapters/

# ====================================================================================

FROM alpine

RUN apk --no-cache add ca-certificates
COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /lxbot /lxbot

WORKDIR /lxbot
VOLUME /lxbot/adapters