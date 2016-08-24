[![Build Status](https://travis-ci.org/retzkek/rabbitbeat.svg?branch=master)](https://travis-ci.org/retzkek/rabbitbeat)

# Rabbitbeat

Rabbitbeat reads messages from an AMQP 0.9.1 broker and indexes them in Elasticsearch or to any of
the other locations supported by [libbeat](https://github.com/elastic/beats/libbeat), including:

* logstash
* kafka
* redis
* file
* stdout

Rabbitbeat is functional, but has not been thoroughly tested.

## Getting Started with Rabbitbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Install

```
go get github.com/retzkek/rabbitbeat
```
