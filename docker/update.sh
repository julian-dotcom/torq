#!/usr/bin/env bash

BASEDIR=$(dirname "$0")

docker-compose -f $BASEDIR/docker-compose.yml down
docker pull lncapital/torq
docker-compose -f $BASEDIR/docker-compose.yml up -d
