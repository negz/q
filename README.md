# q  [![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/negz/q) [![Travis](https://img.shields.io/travis/negz/q.svg?maxAge=300)](https://travis-ci.org/negz/q/) [![Codecov](https://img.shields.io/codecov/c/github/negz/q.svg?maxAge=3600)](https://codecov.io/gh/negz/q/)
A toy in-memory queueing service with a lot of plumbing. q exposes an arbitrary
number of in-memory FIFO queues via gRPC. Each queue supports add, peek, and pop
operations. Queues may be limited in size or unbounded.

Both queues and messages may be tagged. Queue tags may be updated, but message
tags (and messages in general) are immutable.

Of course, with all queues and their messages being stored in-memory, everything
will be forgotten when the `q` process dies. :)

# Components
The q service consists of three binaries:
* `q` - The main logic. Serves a gRPC API on port 10002.
* `qrest` - Serves a (mostly) automatically generated REST to gRPC gateway on port 80. See `rpc/proto/q.swagger.json` for the API spec.
* `qcli` - A commandline gRPC client for `q`.

# Metrics, logging, and management
`q` exposes Prometheus metrics via HTTP at `/metrics` on port 10003. We expose
the count of total enqueued and consumed messages, tagged by queue ID. Total
errors are also exposed, tagged by queue and error type. We only expose counts,
not gauges, because counts
[don't lose meaning when downsampled in a timeseries](https://goo.gl/WTHgAq).

`qrest` also exposes Prometheus metrics at `/metrics` on port 80, but only the
process and Go runtime information Prometheus provides for free.

Both `q` and `qrest` provide terse JSON structured logs on stdout by default.
Run them with the `-d` flag for debug logging in a more human friendly format.

`q` can be terminated immediately by hitting `/quitquitquit` on its metrics
port. `qrest` can also be terminated by hitting `/quitquitquit` on its main
port.

# Packages
`q` consists of the following packages. Refer to their GoDocs for API details:
* [q](https://godoc.org/github.com/negz/q) - Defines the core interfaces and types for the queue service.
* [q/e](https://godoc.org/github.com/negz/q/e) - Provides error types and handling.
* [q/factory](https://godoc.org/github.com/negz/q/factory) - A `q.Factory` implementation.
* [q/logging](https://godoc.org/github.com/negz/q/logging) - Log emitting wrappers for `q.Queue` and `q.Manager`.
* [q/manager](https://godoc.org/github.com/negz/q/manager) - Implementations of `q.Manager`.
* [q/memory](https://godoc.org/github.com/negz/q/memory) - An in-memory implementation of `q.Queue`.
* [q/metrics](https://godoc.org/github.com/negz/q/metrics) - Metric emitting wrappers for `q.Queue`.
* [q/rpc](https://godoc.org/github.com/negz/q/rpc) - Implements gRPC API for `q`.
* [q/proto](https://godoc.org/github.com/negz/q/proto) - Protocol buffer specification for the gRPC API and database serialisation.
* [q/test/fixtures](https://godoc.org/github.com/negz/q/test/fixtures) - Common fixtures used to test `q`.

# Running
Kubernetes deployment configs are provided under `kube/`. Use minikube to run
`q` locally:
```
# Install Minikube and deploy q
$ brew cask install minikube
$ minikube start
$ kubectl create namespace q
$ kubectl -n q create -f kube/deployment.yaml
$ kubectl -n q create -f kube/service.yaml

# Use kubectl to determine which node ports map to which internal ports.
$ kubectl -n q describe service q|grep Port
Type:                   NodePort
Port:                   grpc    10002/TCP
NodePort:               grpc    31051/TCP
Port:                   metrics 10003/TCP
NodePort:               metrics 31457/TCP
Port:                   rest    80/TCP
NodePort:               rest    30647/TCP

$ minikube service -n q q --url
http://192.168.99.101:31051  # gRPC is listening here.
http://192.168.99.101:31457  # Metrics are being served here.
http://192.168.99.101:30647  # The REST gateway is being served here.

# Use it!
$ docker pull negz/qcli
$ docker run negz/qcli /qcli -s 192.168.99.101:31051 new MEMORY 10 -t function="cubesat launcher"
{
  "queue": {
    "meta": {
      "id": "f9e0925d-bfaa-4e59-96ae-dd78a0bb751d",
      "created": "2017-06-25T23:01:08.008101953Z",
      "tags": [
        {
          "key": "function",
          "value": "cubesat launcher"
        }
      ]
    },
    "store": "MEMORY"
  }
}
$ echo "dove"|docker run negz/qcli /qcli -s 192.168.99.101:31051 add f9e0925d-bfaa-4e59-96ae-dd78a0bb751d -t size=3U
{
  "message": {
    "meta": {
      "id": "91432eee-e5ef-4f14-84ab-eb471e2e9d61",
      "created": "2017-06-25T23:05:44.628686920Z",
      "tags": [
        {
          "key": "size",
          "value": "3U"
        }
      ]
    }
  }
}
$ curl -s http://192.168.99.101:31457/metrics|grep queue
# HELP queue_messages_enqueued_total Number of queued messages.
# TYPE queue_messages_enqueued_total counter
queue_messages_enqueued_total{queue="f9e0925d-bfaa-4e59-96ae-dd78a0bb751d"} 1
```

# Building
You'll need working Go and Docker installs. This project has been built and
tested against Go 1.8. Clone the project and run the build script to compile
the binaries and create Docker images:
```
$ mkdir -p ${GOPATH}/src/github.com/negz
$ cd ${GOPATH}/src/github.com/negz
$ git clone git@github.com:negz/q
$ cd q
$ scripts/build.sh
```

# Testing
Take a look at the [Travis CI](https://travis-ci.org/negz/q/) project, or run
`scripts/test.sh`. Note that `test.sh` assumes you've setup your environment
per the build instructions.

# Generated code
Any directory with a `generate.go` file contains automatically generated code
generated by running `go generate`.

This code generation depends on tools that are not managed by dep. To regenerate
code you'll need to:
```
$ go get golang.org/x/tools/cmd/stringer
$ go get github.com/gogo/protobuf/protoc-gen-gogoslick
$ go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
$ go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
```

# Known issues
* The `NewMessage` protocol buffer message is not included in the generated
Swagger API docs. This makes it difficult to discover how to add new messages to
a queue.
