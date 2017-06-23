#!/usr/bin/env sh

# Allow importing proto files using the same pattern as Go imports.
# Special case for grpc-gateway, which attempts a relative import of
# google/api/http.proto
INCLUDE=".:${GOPATH}/src:${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"

protoc -I ${INCLUDE} \
    --gogoslick_out=plugins=grpc:. \
    --grpc-gateway_out=logtostderr=true,allow_delete_body=true:. \
    --swagger_out=logtostderr=true,allow_delete_body=true:. \
    q.proto