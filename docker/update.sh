#!/usr/bin/env bash

BASEDIR=$(dirname "$0")

docker-compose -f $BASEDIR/docker-compose.yml down
docker rm $(docker ps -aq  --filter "ancestor=lncapital/torq:latest")
docker rmi lncapital/torq:latest
docker-compose -f $BASEDIR/docker-compose.yml up -d
