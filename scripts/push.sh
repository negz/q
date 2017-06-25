#!/usr/bin/env bash

set -e

docker login -u="${DOCKER_USERNAME}" -p="${DOCKER_PASSWORD}"

docker push negz/queue:$(git rev-parse --short HEAD)
docker push negz/queue:latest

docker push negz/qrest:$(git rev-parse --short HEAD)
docker push negz/qrest:latest

docker push negz/qcli:$(git rev-parse --short HEAD)
docker push negz/qcli:latest
