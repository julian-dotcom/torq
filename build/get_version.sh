#!/bin/bash
TAG=`git describe --tags --abbrev=0`
HASH=`git rev-parse HEAD`
DATE=`date`
echo -n "${TAG} | ${HASH} | ${DATE}" > build/version.txt

