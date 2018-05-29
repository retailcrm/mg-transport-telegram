FROM golang:1.9.3-stretch

WORKDIR /
ADD ./bin/mg-telegram /
ADD ./templates/ /templates/
ADD ./static/ /static/
ADD ./translate/ /translate/
ADD ./migrations/ /migrations/

EXPOSE 3001

ENTRYPOINT ["/mg-telegram"]

CMD ["run"]
