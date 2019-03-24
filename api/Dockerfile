# Build Binary
FROM golang:alpine as build
RUN apk add make ca-certificates
RUN mkdir -p $GOPATH/src/github.com/hackathon/hackhub
WORKDIR $GOPATH/src/github.com/hackathon/hackhub
ADD ./ ./
RUN make && cp hackhub /

# Inject
FROM scratch
ADD config.yaml ./
COPY --from=build /hackhub /bin/hackhub
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["hackhub"]
