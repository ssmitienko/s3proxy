# build stage
FROM golang:1.22.1-alpine3.19 AS build-env

RUN mkdir -p /src/
ADD . /src
RUN cd /src && go build

# final stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates && update-ca-certificates

COPY --from=build-env /src/s3proxy /s3proxy
COPY proxy.json /proxy.json

EXPOSE 8123
STOPSIGNAL SIGTERM
LABEL maintainer="Serhii Smitiienko <sergey.smitienko@gmail.com>"

CMD ["/s3proxy", "-listen", "0.0.0.0:8123", "-config", "/proxy.json"]

