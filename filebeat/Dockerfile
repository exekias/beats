FROM ubuntu:16.04
MAINTAINER Elastic Docker Team <docker@elastic.co>

RUN apt-get update
RUN apt-get install -y ca-certificates

RUN mkdir -p /usr/share/filebeat

COPY filebeat /usr/share/filebeat

COPY filebeat.yml /usr/share/filebeat/
COPY fields.yml /usr/share/filebeat/
COPY module /usr/share/filebeat/module

WORKDIR /usr/share/filebeat

ENTRYPOINT ["./filebeat", "-e"]
