FROM alpine:3.20 AS base
RUN apk add --no-cache ca-certificates tzdata
COPY dbxcli /usr/local/bin/dbxcli
COPY dbxctl /usr/local/bin/dbxctl
EXPOSE 8080
ENTRYPOINT ["dbxcli"]
CMD ["--help"]
