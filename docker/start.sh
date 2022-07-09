#!/usr/bin/env bash

BASEDIR=$(dirname "$0")

docker pull lncapital/torq
docker-compose -f $BASEDIR/docker-compose.yml up  -d

echo Torq is starting, please wait

function timeout() { perl -e 'alarm shift; exec @ARGV' "$@"; }

timeout 300 bash -c 'while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:8080)" != "200" ]]; do sleep 5; done' || false

echo Torq has started and is available on http://localhost:8080
if [ "$(uname)" == "Darwin" ]; then
    open http://localhost:8080
fi
if [[ "$(uname)" != "Darwin" && x$DISPLAY != x ]]; then
  xdg-open http://localhost:8080
fi
