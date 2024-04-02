FROM golang:alpine3.18 as builder
WORKDIR /app
COPY . .
RUN apk --no-cache --update add git ca-certificates  build-base
# RUN apk add --no-cache --update git build-base   
RUN  CGO_ENABLED=1 go build -o main -trimpath -ldflags "-s -w -buildid=" ./cmd/


FROM alpine:3.18 as runner
RUN apk --no-cache add ca-certificates tzdata
ENV LANG C.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL C.UTF-8
WORKDIR /app

ENTRYPOINT ["./main"]

EXPOSE 8082
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/main /app/main
