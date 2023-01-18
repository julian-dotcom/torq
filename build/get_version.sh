#!/bin/bash
echo "GITHUB_REPOSITORY=${GITHUB_REPOSITORY}"
echo "GITHUB_REF_NAME=${GITHUB_REF_NAME}"
echo "GITHUB_SHA=${GITHUB_SHA}"
TAG=`git describe --tags --abbrev=0`
HASH=`git rev-parse HEAD`
DATE=`date`
echo "TAG=${TAG}"
echo "HASH=${HASH}"
echo "DATE=${DATE}"
echo -n "${TAG} | ${HASH} | ${DATE}" > version.txt
