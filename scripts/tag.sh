#!/bin/bash
echo "****************************************"
echo "*** Tag docker image with the branch name  ***"
echo "****************************************"

# Docker image name on docker hub
IMAGE_NAME=consensys/fc-retrieval-provider

VERSION=$1
echo "Register version: $VERSION"
IMAGE=$2
echo "v image: $IMAGE"


ITEST_BRANCH=`git rev-parse --abbrev-ref HEAD`
echo "Provider repo branch: $ITEST_BRANCH"

TAG="develop-$ITEST_BRANCH"
echo "TAG: $TAG"

docker tag $IMAGE $IMAGE_NAME:$TAG
