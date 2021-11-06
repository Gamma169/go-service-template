#!/bin/bash

# A script to quickly build & publish docker images defined in this directory

# usage:
# ./build.sh [command] [image]

# examples:
# ./build.sh build              // builds foobar image
# ./build.sh publish            // publishes foobar image
# ./build.sh build_and_publish  // builds & publishes foobar image

# will exit out if something fails
set -e
set -o pipefail

DOCKER_ORG="gamma169"
SERVICE_NAME="foobar"

DOCKER_TAG=${DOCKER_TAG:-"latest"}
DOCKER_USERNAME=$DOCKER_USERNAME
DOCKER_PASSWORD=$DOCKER_PASSWORD

# the default location for a service's Dockerfile
DOCKERFILE_LOCATION="Dockerfile"

# build the docker image.
function build {

   # Pre-build hook.
  if [ -e "deploy/pre_build.sh" ]
    then
      ./deploy/pre_build.sh
  fi

  echo "docker build -t $DOCKER_IMAGE_AND_TAG -f $DOCKERFILE_LOCATION ."
  docker build -t "$DOCKER_IMAGE_AND_TAG" -f $DOCKERFILE_LOCATION .

  # Cleanup script
  if [ -e "deploy/cleanup.sh" ]
    then
      ./deploy/cleanup.sh
  fi
}

# publish the docker image to the docker registry.
function publish {
  docker login -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"
  docker push "$DOCKER_IMAGE_AND_TAG"
}

function build_and_publish {
  build
  publish
}

if [[ $1 =~ ^(build|publish|build_and_publish)$ ]];
then
  DOCKER_IMAGE_AND_TAG="$DOCKER_ORG/$SERVICE_NAME:$DOCKER_TAG"
  "$@"
else
  echo "Argument '$1' is not a valid command."
  exit 1
fi
