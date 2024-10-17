#!/bin/sh

export DOCKER_BUILDKIT=1
set -e

IMAGE="zweknu/spellingbee"
if [ "$#" != 1 ]; then
  echo 1>&2 "Usage $0 <tag>"
  exit 1
fi
TAG=$1

echo "Building $IMAGE:$TAG..."
go test
docker buildx build --push -t $IMAGE:$TAG .
## echo "Pushing..."
## docker push $IMAGE:$TAG
# kubectl set image deployments/spellingbee-grpc spellingbee-server=docker.io/$IMAGE:$TAG
echo "Done."
