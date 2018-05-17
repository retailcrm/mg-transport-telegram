FROM golang:1.9-stretch as ca-certs
FROM scratch

COPY --from=ca-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /

EXPOSE 3001

ENTRYPOINT ["/mg-transport-telegram", "--config", "/config.yml"]

CMD ["run"]
