FROM golang:1.9.3-stretch

WORKDIR /
ADD ./bin/mg-telegram /
ADD ./templates/ /templates/
ADD ./web/ /web/
ADD ./translate/ /translate/

EXPOSE 3001

ENTRYPOINT ["/mg-telegram"]

CMD ["run"]
