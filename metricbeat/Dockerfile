FROM ubuntu:16.04
MAINTAINER Elastic Docker Team <docker@elastic.co>

RUN apt-get update
RUN apt-get install -y ca-certificates

RUN mkdir -p /usr/share/metricbeat

COPY metricbeat /usr/share/metricbeat/

COPY metricbeat.yml /usr/share/metricbeat/
COPY fields.yml /usr/share/metricbeat/

WORKDIR /usr/share/metricbeat

ENTRYPOINT ["./metricbeat", "-e"]
