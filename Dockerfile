FROM golang:1.9.3-stretch

WORKDIR /
ADD ./bin/mg-telegram /
ADD ./templates/ /templates/

EXPOSE 3001

ENTRYPOINT ["/mg-telegram", "run"]
