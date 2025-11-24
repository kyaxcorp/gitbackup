#!/bin/bash

REPO="gitbackup/gitbackup"
TAG_NAME="$REPO:v1.0.5"
LATEST="$REPO:latest"

docker buildx create --use --name mounted-build-kit --node mounted-build-kit --bootstrap
if ! docker buildx build \
     --builder mounted-build-kit \
     --push \
     --platform linux/amd64,linux/arm64 \
     --tag $TAG_NAME --tag $LATEST -f Dockerfile . ; then
  echo "build error, terminating script"
  exit 1
fi
docker buildx prune -a -f --builder mounted-build-kit

echo "build completed successfully"

#if ! docker build --compress --tag $TAG_NAME --tag $LATEST -f Dockerfile .; then
#  echo "build error, terminating script"
#  exit 1
#fi
#echo "build completed successfully"

#docker push $TAG_NAME
#docker push $LATEST